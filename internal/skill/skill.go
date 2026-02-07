package skill

type Skill struct {
	Name        string
	Description string
	AbsPath     string
	Path        string
	Content     string
	Body        string
	Hash        string
}

type SkillList struct {
	ByName map[string]*Skill
	ByPath map[string]*Skill
	Paths  []string
}

func NewSkillList() *SkillList {
	return &SkillList{
		ByName: make(map[string]*Skill),
		ByPath: make(map[string]*Skill),
	}
}

func (s *SkillList) List() []string {
	names := make([]string, 0, len(s.ByName))
	for name := range s.ByName {
		names = append(names, name)
	}
	return names
}
