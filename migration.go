package migrations

import (
	"bufio"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
	"fmt"
)

const (
	MigrationTimeLayout = "20060102150405"
)

type Migration struct {
	Name    string
	Version time.Time

	Content struct {
		Up   string
		Down string
	}
}

func (m *Migration) VersionAsString() string {
	return m.Version.Format(MigrationTimeLayout)
}

type Migrations []Migration

func (m Migrations) Len() int           { return len(m) }
func (m Migrations) Less(i, j int) bool { return m[i].Version.Before(m[j].Version) }
func (m Migrations) Swap(i, j int)      { m[i], m[j] = m[j], m[i] }

func (m Migrations) Up(driver Driver) {
	sort.Reverse(m)

	driver.CreateVersionsTable()

	for _, migration := range m {
		if driver.HasExecuted(migration.VersionAsString()) {
			continue
		}

		fmt.Println(migration.Content.Up)
		driver.Up(migration)
	}
}

func (m Migrations) Down(driver Driver) {
	sort.Sort(m)

	driver.CreateVersionsTable()

	for _, migration := range m {
		if !driver.HasExecuted(migration.VersionAsString()) {
			continue
		}

		fmt.Println(migration.Content.Up)
		driver.Down(migration)
		break
	}
}

func CreateFromDirectory(dir string) Migrations {

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		panic(err)
	}

	migrations := Migrations{}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if strings.HasSuffix(file.Name(), ".sql") {
			migrations = append(migrations, newMigrationFromPath(path.Join(dir, file.Name())))
		}
	}

	return migrations
}

func newMigrationFromPath(path string) Migration {

	baseName := filepath.Base(path)
	unparsedVersion := regexp.MustCompile("^\\d+").FindString(baseName)

	version, err := time.Parse(MigrationTimeLayout, unparsedVersion)
	if err != nil {
		panic(err)
	}

	migration := Migration{
		Name:    baseName,
		Version: version,
	}

	reader, err := os.Open(path)
	if err != nil {
		panic(err)
	}

	scanner := bufio.NewScanner(reader)

	up := true

	for scanner.Scan() {
		line := scanner.Text()

		switch true {
		case strings.HasPrefix(line, "-- up"):
			up = true

		case strings.HasPrefix(line, "-- down"):
			up = false

		default:
			if up {
				migration.Content.Up += line + "\n"
			} else {
				migration.Content.Down += line + "\n"
			}
		}
	}

	return migration
}
