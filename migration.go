package migrations

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/pkg/errors"
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

func (m Migrations) Up(driver Driver, verbose bool) error {
	if err := driver.CreateVersionsTable(); err != nil {
		return err
	}

	for _, migration := range m {
		if driver.HasExecuted(migration.VersionAsString()) {
			continue
		}

		if verbose {
			fmt.Println(migration.Content.Up)
		}

		if err := driver.Up(migration); err != nil {
			return err
		}
	}

	return nil
}

func (m Migrations) Down(driver Driver, verbose bool) error {
	return errors.New("Dont use this, this project is archived")
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

	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}

	defer func() {
		_ = file.Close()
	}()

	scanner := bufio.NewScanner(file)

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
