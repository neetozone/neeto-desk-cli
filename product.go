package product

import _ "embed"

//go:embed .neeto-cli.yml
var ConfigYAML []byte

//go:embed skills/neetodesk/SKILL.md
var SkillMD []byte
