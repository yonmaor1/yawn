// Package shell spawns an interactive subshell for `yawn switch`: it cd's into
// the task's work directory, applies the task's rc commands, shows a [name]
// prompt prefix, and runs the cleanup commands when the shell exits.
//
// rc and cleanup injection are supported for bash and zsh (which covers macOS's
// default zsh and bash). Other shells get a plain interactive shell in the work
// directory with $YAWN_TASK set, and a note that rc/cleanup were skipped.
package shell

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Run launches the subshell and blocks until it exits. Non-zero exit codes from
// the user's shell are not treated as errors; only a failure to start is.
func Run(workdir, taskName string, rc, cleanup []string) error {
	shellPath := os.Getenv("SHELL")
	if shellPath == "" {
		shellPath = "/bin/sh"
	}
	base := filepath.Base(shellPath)

	var cmd *exec.Cmd
	var cleanupFn func()

	switch base {
	case "bash":
		rcfile, err := writeTemp("yawn-bashrc", bashInit(workdir, taskName, rc, cleanup))
		if err != nil {
			return err
		}
		cleanupFn = func() { os.Remove(rcfile) }
		cmd = exec.Command(shellPath, "--rcfile", rcfile, "-i")
	case "zsh":
		dir, err := os.MkdirTemp("", "yawn-zdotdir")
		if err != nil {
			return err
		}
		cleanupFn = func() { os.RemoveAll(dir) }
		if err := os.WriteFile(filepath.Join(dir, ".zshrc"), []byte(zshInit(workdir, taskName, rc, cleanup)), 0o644); err != nil {
			cleanupFn()
			return err
		}
		cmd = exec.Command(shellPath, "-i")
		cmd.Env = append(os.Environ(), "ZDOTDIR="+dir)
	default:
		if len(rc) > 0 || len(cleanup) > 0 {
			fmt.Fprintf(os.Stderr, "yawn: rc/cleanup are only applied for bash and zsh; skipped for %s\n", base)
		}
		cmd = exec.Command(shellPath, "-i")
	}

	if cleanupFn != nil {
		defer cleanupFn()
	}
	if cmd.Env == nil {
		cmd.Env = os.Environ()
	}
	cmd.Env = append(cmd.Env, "YAWN_TASK="+taskName)
	cmd.Dir = workdir
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if _, ok := err.(*exec.ExitError); ok {
		return nil // user exited with a non-zero status; not a yawn error
	}
	return err
}

func writeTemp(prefix, content string) (string, error) {
	f, err := os.CreateTemp("", prefix)
	if err != nil {
		return "", err
	}
	if _, err := f.WriteString(content); err != nil {
		f.Close()
		os.Remove(f.Name())
		return "", err
	}
	return f.Name(), f.Close()
}

func bashInit(workdir, name string, rc, cleanup []string) string {
	var b strings.Builder
	b.WriteString("[ -f \"$HOME/.bashrc\" ] && source \"$HOME/.bashrc\"\n")
	fmt.Fprintf(&b, "cd %s\n", quote(workdir))
	for _, c := range rc {
		b.WriteString(c + "\n")
	}
	fmt.Fprintf(&b, "PS1=%s\"$PS1\"\n", quote("["+name+"] "))
	if len(cleanup) > 0 {
		b.WriteString("__yawn_cleanup() {\n")
		for _, c := range cleanup {
			b.WriteString("  " + c + "\n")
		}
		b.WriteString("}\n")
		b.WriteString("trap __yawn_cleanup EXIT\n")
	}
	return b.String()
}

func zshInit(workdir, name string, rc, cleanup []string) string {
	var b strings.Builder
	b.WriteString("[ -f \"$HOME/.zshrc\" ] && source \"$HOME/.zshrc\"\n")
	fmt.Fprintf(&b, "cd %s\n", quote(workdir))
	for _, c := range rc {
		b.WriteString(c + "\n")
	}
	fmt.Fprintf(&b, "PROMPT=%s$PROMPT\n", quote("["+name+"] "))
	if len(cleanup) > 0 {
		b.WriteString("function zshexit() {\n")
		for _, c := range cleanup {
			b.WriteString("  " + c + "\n")
		}
		b.WriteString("}\n")
	}
	return b.String()
}

// quote single-quotes a string for safe use in a shell script.
func quote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}
