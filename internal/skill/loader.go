// AstrCode - Agent orchestration engine for AstrBot
// Copyright (C) 2026 EterUltimate
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

package skill

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/EterUltimate/astrcode/internal/model"
)

// Loader 负责从文件系统加载 Skill
type Loader struct {
	basePath string
}

// NewLoader 创建新的 Skill 加载器
func NewLoader(basePath string) *Loader {
	return &Loader{basePath: basePath}
}

// LoadFromDirectory 从目录加载所有 Skill
func (l *Loader) LoadFromDirectory() ([]model.Skill, error) {
	entries, err := os.ReadDir(l.basePath)
	if err != nil {
		return nil, fmt.Errorf("read skill directory: %w", err)
	}

	var skills []model.Skill
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		skill, err := l.loadSkill(entry.Name())
		if err != nil {
			continue // 跳过加载失败的 Skill
		}
		skills = append(skills, skill)
	}

	return skills, nil
}

// loadSkill 加载单个 Skill
func (l *Loader) loadSkill(name string) (model.Skill, error) {
	skillPath := filepath.Join(l.basePath, name)

	// 读取 SKILL.md
	skillFile := filepath.Join(skillPath, "SKILL.md")
	content, err := os.ReadFile(skillFile)
	if err != nil {
		return model.Skill{}, fmt.Errorf("read SKILL.md: %w", err)
	}

	// 解析 SKILL.md
	skill := parseSkillMarkdown(name, string(content))
	skill.Path = skillPath

	return skill, nil
}

// parseSkillMarkdown 解析 SKILL.md 内容
func parseSkillMarkdown(name, content string) model.Skill {
	skill := model.Skill{
		Name: name,
	}

	lines := strings.Split(content, "\n")
	var inDescription bool
	var descLines []string

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// 提取标题作为描述
		if strings.HasPrefix(line, "# ") && skill.Description == "" {
			skill.Description = strings.TrimPrefix(line, "# ")
			inDescription = true
			continue
		}

		// 收集描述段落
		if inDescription && line != "" && !strings.HasPrefix(line, "#") {
			descLines = append(descLines, line)
		} else if inDescription && strings.HasPrefix(line, "#") {
			inDescription = false
		}
	}

	if len(descLines) > 0 {
		skill.Description = strings.Join(descLines, " ")
	}

	return skill
}
