package skill

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
)

type Scanner struct {
	paths  []string
	skills *SkillList
	mu     sync.RWMutex
}

func NewScanner() *Scanner {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}

	cwd, _ := os.Getwd()
	paths := []string{
		filepath.Join(cwd, ".claude", "skills"),
		filepath.Join(cwd, ".skills"),
		// Claude
		filepath.Join(home, ".claude", "skills"),
		// OpenCode
		filepath.Join(home, ".opencode", "skills"),
		// OpenAI / Codex
		filepath.Join(home, ".openai", "skills"),
		filepath.Join(home, ".codex", "skills"),
		"/mnt/skills/public",
		"/mnt/skills/user",
		"/mnt/skills/examples",
	}

	return &Scanner{
		paths:  paths,
		skills: NewSkillList(),
	}
}

func (l *Scanner) Scan() (*SkillList, error) {
	list := NewSkillList()
	list.Paths = l.paths

	// * concurrent scan path list
	var wg sync.WaitGroup
	skillChan := make(chan *Skill, 100)
	errChan := make(chan error, len(l.paths))
	for _, path := range l.paths {
		wg.Add(1)

		go func(dir string) {
			defer wg.Done()
			if err := l.scan(dir, skillChan); err != nil {
				errChan <- fmt.Errorf("failed to scan %s: %w", dir, err)
			}
		}(path)
	}

	go func() {
		wg.Wait()
		close(skillChan)
		close(errChan)
	}()

	for skill := range skillChan {
		if _, ok := list.ByName[skill.Name]; ok {
			continue
		}
		list.ByName[skill.Name] = skill
		list.ByPath[skill.AbsPath] = skill
	}

	var errs []error
	for err := range errChan {
		errs = append(errs, err)
		slog.Warn("scan error", "error", err)
	}

	l.mu.Lock()
	l.skills = list
	l.mu.Unlock()

	return list, nil
}

func (l *Scanner) scan(root string, skillChan chan<- *Skill) error {
	// * path not exists
	if _, err := os.Stat(root); os.IsNotExist(err) {
		return nil
	}

	entries, err := os.ReadDir(root)
	if err != nil {
		return err
	}

	for _, e := range entries {
		if !e.IsDir() || e.Name()[0] == '.' {
			continue
		}

		// ~/.claude/skills/
		// └── {skill_name}/
		//     └── SKILL.md
		path := filepath.Join(root, e.Name(), "SKILL.md")
		if _, err := os.Stat(path); os.IsNotExist(err) {
			continue
		}

		skill, err := parse(path)
		if err != nil {
			continue
		}
		skillChan <- skill
	}

	return nil
}
