package yaml

import (
	"testing"
)

var data = `
name: Example Developer
job: Developer
skill: Elite
employed: True
foods:
    - Apple
    - Orange
    - Strawberry
    - Mango
languages:
    ruby: Elite
    python: Elite
    dotnet: Lame
`

type RollingFileConfig struct {
	FileDir   string `yaml:"FileDir"`
	FileName  string `yaml:"FileName"`
	MaxNumber int32  `yaml:"MaxNumber"`
	MaxSize   int64  `yaml:"MaxSize"`
	Unit      int64  `yaml:"Unit"`
}

type RollingDailyConfig struct {
	FileDir  string `yaml:"FileDir"`
	FileName string `yaml:"FileName"`
}

type LogYaml struct {
	IsConsole    bool
	Level        int32
	Format       int
	RollingFile  RollingFileConfig
	RollingDaily RollingDailyConfig
}

type T struct {
	Name      string
	Job       string
	Skill     string
	Employed  bool
	Foods     []string
	Languages map[string]string
}

func TestUnmarshal(t *testing.T) {
	value := T{}

	if err := Unmarshal([]byte(data), &value); err != nil {
		t.Fatalf("error: %v", err)
	}
	if value.Employed == true {
		t.Log("passed")
	} else {
		t.Log("failed")
	}
}

func TestMarshal(t *testing.T) {
	value := T{
		Name:      "longwu",
		Job:       "Developer",
		Skill:     "Golang",
		Employed:  true,
		Foods:     []string{"apple", "pea"},
		Languages: map[string]string{"english": "UK"},
	}

	if _, err := Marshal(&value); err != nil {
		t.Fatalf("error: %v", err)
	} else {
		t.Log("passed")
	}
}

func TestMarshalToFile(t *testing.T) {
	value := T{
		Name:      "longwu",
		Job:       "Developer",
		Skill:     "Golang",
		Employed:  true,
		Foods:     []string{"apple", "pea"},
		Languages: map[string]string{"english": "UK"},
	}

	name := "job.yaml"

	if err := MarshalToFile(&value, name); err != nil {
		t.Fatalf("error: %v", err)
	} else {
		t.Log("passed")
	}
}

func TestMarshalLogYaml(t *testing.T) {

	type LogYaml struct {
		IsConsole    bool               `yaml:"IsConsole"`
		Level        int32              `yaml:"Level"`
		Format       int                `yaml:"Format"`
		RollingFile  RollingFileConfig  `yaml:"RollingFile"`
		RollingDaily RollingDailyConfig `yaml:"RollingDaily"`
	}
	log := LogYaml{
		IsConsole: false,
		Level:     2,
		Format:    3,
		RollingFile: RollingFileConfig{
			FileDir:   "/tmp",
			FileName:  "test.log",
			MaxNumber: 5,
			MaxSize:   10000000,
			Unit:      4,
		},
		RollingDaily: RollingDailyConfig{
			FileDir:  "/tmp/daily",
			FileName: "daily.log",
		},
	}

	name := "log.yaml"

	if err := MarshalToFile(&log, name); err != nil {
		t.Fatalf("error: %v", err)
	} else {
		t.Log("passed")
	}
}

func TestUnmarshalFromFile(t *testing.T) {
	name := "job.yaml"

	value := T{}
	if err := UnmarshalFromFile(name, &value); err != nil {
		t.Fatalf("error: %v", err)
	} else {
		if value.Employed == true {
			// t.Log(value)
			t.Log("passed")
		} else {
			t.Log("failed")
		}
	}
}

func TestUnmarshalLog(t *testing.T) {
	v := `
IsConsole: false
Level: 2
Format: 3
RollingFile:
  FileDir: /tmp
  FileName: test.log
  MaxNumber: 5
  MaxSize: 10000000
  Unit: 4
#RollingDaily:
#  FileDir: /tmp/daily
#  FileName: daily.log
`
	type LogYaml struct {
		IsConsole    bool               `yaml:"IsConsole"`
		Level        int32              `yaml:"Level"`
		Format       int                `yaml:"Format"`
		RollingFile  RollingFileConfig  `yaml:"RollingFile"`
		RollingDaily RollingDailyConfig `yaml:"RollingDaily"`
	}
	log := LogYaml{}
	err := Unmarshal([]byte(v), &log)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if log.RollingDaily.FileDir != "" {
		t.Fatalf("daily")
	}

	if log.RollingFile.FileDir != "/tmp" {
		t.Fatalf("rolling")
	}
}
