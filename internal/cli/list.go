package cli

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yonmaor/yawn/internal/config"
	"github.com/yonmaor/yawn/internal/repo"
	"github.com/yonmaor/yawn/internal/state"
)

var (
	listIn    string
	listToday bool
)

var listCmd = &cobra.Command{
	Use:   "list [--in <name>] [--today]",
	Short: "list your tasks",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := requireRepo()
		if err != nil {
			return err
		}
		tasks, err := loadTasks(r)
		if err != nil {
			return err
		}
		if len(tasks) == 0 {
			fmt.Println("no tasks yet — `yawn new <name>`")
			return nil
		}

		switch {
		case listToday:
			return listTodayTasks(r, tasks)
		case listIn != "":
			return listChildren(tasks, listIn)
		default:
			printTree(tasks)
			return nil
		}
	},
}

func init() {
	listCmd.Flags().StringVar(&listIn, "in", "", "only tasks under this parent")
	listCmd.Flags().BoolVar(&listToday, "today", false, "only tasks in today's todo")
}

func listTodayTasks(r *repo.Repo, tasks map[string]*config.Task) error {
	s, err := state.Load(r.Dir)
	if err != nil {
		return err
	}
	names := s.Get(today())
	if len(names) == 0 {
		fmt.Printf("nothing prioritized for %s\n", today())
		return nil
	}
	for _, n := range names {
		if t, ok := tasks[n]; ok {
			fmt.Println(line(n, t))
		} else {
			fmt.Printf("%s [missing]\n", n)
		}
	}
	return nil
}

func listChildren(tasks map[string]*config.Task, parent string) error {
	var names []string
	for n, t := range tasks {
		if t.Parent == parent {
			names = append(names, n)
		}
	}
	if len(names) == 0 {
		fmt.Printf("no tasks under %q\n", parent)
		return nil
	}
	sort.Strings(names)
	for _, n := range names {
		fmt.Println(line(n, tasks[n]))
	}
	return nil
}

// printTree renders tasks as a hierarchy by parent, rooted at the base branch.
func printTree(tasks map[string]*config.Task) {
	children := map[string][]string{}
	for n, t := range tasks {
		p := t.Parent
		if p == "" {
			p = repo.BaseBranch
		}
		children[p] = append(children[p], n)
	}
	for _, c := range children {
		sort.Strings(c)
	}

	printed := map[string]bool{}
	var walk func(parent string, depth int)
	walk = func(parent string, depth int) {
		for _, n := range children[parent] {
			if printed[n] {
				continue
			}
			printed[n] = true
			fmt.Printf("%s%s\n", strings.Repeat("  ", depth), line(n, tasks[n]))
			walk(n, depth+1)
		}
	}
	walk(repo.BaseBranch, 0)

	// Any task whose parent is not reachable from the base (orphan) — list flat.
	var orphans []string
	for n := range tasks {
		if !printed[n] {
			orphans = append(orphans, n)
		}
	}
	sort.Strings(orphans)
	for _, n := range orphans {
		fmt.Println(line(n, tasks[n]))
	}
}

func line(name string, t *config.Task) string {
	status := t.Status
	if status == "" {
		status = "?"
	}
	return fmt.Sprintf("%s [%s]", name, status)
}
