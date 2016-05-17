package main

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/fatih/color"
	"github.com/jessevdk/go-flags"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

const VERSION = "0.1.2"

/*
  GitOptions
  Structure handeling command line options related to Git url and branch.
*/
type GitOptions struct {
	GitUrl string `short:"g" long:"git" description:"Git repository to clone"`
	Branch string `short:"b" long:"branch" description:"Branch to work with" default:"master"`
}

/*
  Options
  Structure handeling command line options related to debug and version.
*/
type Options struct {
	Git     GitOptions
	Debug   bool `short:"d" long:"debug" description:"Run in debug mode"`
	Version bool `short:"v" long:"version" description:"Display program version"`
}

/*
  Config
  Structure mirroring the format of a valid .cappuccino.yml file.
  Consists of an engine name, associated version and an array of actions.
*/
type Config struct {
	Engine  string
	Version string
	Actions []Action
}

/*
  Action
  Structure mirroring the format of a valid action if a config file.
  Consists of a name and an array of action commands.
*/
type Action struct {
	Name    string
	Type    string
	Content []ActionContent
}

/*
  ActionContent
  Structure mirroring the format of a valid command if a config file.
  Consists of a type, path, command, source, destination, variable,
  path and a value.
*/
type ActionContent struct {
	Type        string
	Command     string
	Source      string
	Destination string
	Variable    string
	Text        string
	Path        string
	Value       string
	Indent      int
}

func main() {
	var opts Options
	var parser = flags.NewParser(&opts, flags.Default)

	if _, err := parser.Parse(); err != nil {
		os.Exit(0)
	}

	if opts.Version {
		fmt.Printf("%s %s\n", os.Args[0], VERSION)
	}

	gitUrl := opts.Git.GitUrl
	gitBranch := opts.Git.Branch

	if gitUrl != "" {
		startEngine()
		verifyGitUrl(gitUrl)
		cloneRepo(gitUrl, gitBranch)
		unmarshalConfig(gitUrl)
	}
}

/*
  verifyGitUrl
  Checks that the git url given in parameter is in a valid format.
  Exists the program otherwize displaying an error message.
  Same logic should be applied for a SVN cloning process.
*/
func verifyGitUrl(href string) {
	text(fmt.Sprintf("Checking git url format (%s)", href), color.FgYellow)
	regex := "((git|ssh|http(s)?)|(git@[\\w\\.]+))(:(//)?)([\\w\\.@\\:/\\-~]+)(\\.git)(/)?"

	match, _ := regexp.MatchString(regex, href)
	if match {
		content := fmt.Sprintf("Git url format successfuly verified")
		text(content, color.FgYellow)
	} else {
		content := fmt.Sprintf("Git url format is not valid")
		text(content, color.FgRed)
		os.Exit(0)
	}
}

/*
  cloneRepo
  Clones a specific branch of git repository.
  Same logic could be applied for an SVN repo.
*/
func cloneRepo(href string, branch string) {
	content := fmt.Sprintf("Cloning git repository (branch: %s)", branch)
	text(content, color.FgYellow)
	executeCommand("git", "clone", href, "-b", branch)
}

/*
  unmarshalConfig
  Reads the .cappuccino.yml config file and extracts content
  to the Config struct defined above. One extracted, content is parsed
  and falls through the execution process.
*/
func unmarshalConfig(href string) {
	os.Chdir(findRepoName(href))

	config := Config{}
	content, err := ioutil.ReadFile(".cappuccino.yml")

	if err != nil {
		text("Error opening .cappuccino.yml file", color.FgRed)
		os.Exit(0)
	}

	text("File .cappuccino.yml detected", color.FgYellow)

	if err := yaml.Unmarshal(content, &config); err != nil {
		text(err.Error(), color.FgRed)
		os.Exit(0)
	}

	displayVersion(&config)
	processConfig(&config)
	processWarnings()
}

/*
  processConfig
  Takes a Config pointer in argument and loops through the list
  of actions and commands, executing one after another in a
  thread safe executeCommand function.
*/
func processConfig(config *Config) {
	text("Starting execution of actions", color.FgYellow)
	removeGitDirectory()

	for i := 0; i < len(config.Actions); i++ {
		processAction(&config.Actions[i])
	}
}

func processAction(action *Action) {
	text(action.Name, color.FgGreen)

	for j := 0; j < len(action.Content); j++ {
		processContent(action, &action.Content[j])
	}
}

/*
  processContent
  Takes an ActionContent as a parameter and handles the execution
  of action depending of it's type.
*/
func processContent(action *Action, content *ActionContent) {
	var contentType string

	if content.Type == "" {
		contentType = action.Type
	} else {
		contentType = content.Type
	}

	switch contentType {
	case "exec":
		command := content.Command
		coloredContent := fmt.Sprintf("\t-> %s", command)
		text(coloredContent, color.FgGreen)

		executableCommand := strings.Split(command, " ")
		executeCommand(executableCommand[0], executableCommand[1:]...)

	case "replace", "substitute":
		path := content.Path
		value := strings.TrimSpace(content.Value)
		indent := content.Indent

		var variable string
		if contentType == "replace" {
			variable = content.Text
		} else {
			variable = fmt.Sprintf("[cappuccino-var-%s]", content.Variable)
		}

		var shownPath string
		if content.Path != "" {
			shownPath = content.Path
		} else {
			shownPath = "all files"
		}

		coloredName := colored(variable, color.FgCyan)
		coloredContent := fmt.Sprintf("\t-> %s in %s", coloredName, shownPath)
		text(coloredContent, color.FgGreen)

		if err := substituteFile(&path, &variable, &value, &indent); err != nil {
			text(err.Error(), color.FgRed)
			os.Exit(0)
		}

	case "copy", "template":
		var source, destination string

		if contentType == "copy" {
			source = content.Source
			destination = content.Destination
		} else {
			source = fmt.Sprintf(".cappuccino/%s", content.Path)
			destination = content.Path
		}

		coloredSource := colored(source, color.FgBlue)
		coloredDestination := colored(destination, color.FgBlue)
		coloredContent := fmt.Sprintf("\t-> %s -> %s", coloredSource, coloredDestination)
		text(coloredContent, color.FgGreen)

		if err := copyFile(source, destination); err != nil {
			text(err.Error(), color.FgRed)
			os.Exit(0)
		}

	case "move":
		source := content.Source
		destination := content.Destination

		coloredSource := colored(source, color.FgMagenta)
		coloredDestination := colored(destination, color.FgMagenta)
		coloredContent := fmt.Sprintf("\t-> %s -> %s", coloredSource, coloredDestination)
		text(coloredContent, color.FgGreen)

		if err := moveFile(source, destination); err != nil {
			text(err.Error(), color.FgRed)
			os.Exit(0)
		}

	case "delete":
		path := content.Path

		coloredSource := colored(path, color.FgRed)
		coloredContent := fmt.Sprintf("\t-> %s", coloredSource)
		text(coloredContent, color.FgGreen)

		if err := deleteFile(path); err != nil {
			text(err.Error(), color.FgRed)
			os.Exit(0)
		}
	}
}

/*
  executeCommand
  Executes a kernel thread safe command with associated arguments
  defined as a vector of infinite sub-components. This displays the
  stdout in case the debug mode is enabled, and omit otherwize.
*/
func executeCommand(command string, args ...string) {
	cmd := exec.Command(command, args...)

	var out bytes.Buffer
	var stderr bytes.Buffer

	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		text(stderr.String(), color.FgRed)
		os.Exit(0)
	}

	if out.String() != "" && false {
		text(out.String(), color.FgGreen)
	}
}

/*
  removeGitDirectory
  Removes the `.git` directory after clone, by default.
*/
func removeGitDirectory() {
	text("Removing existing .git folder", color.FgGreen)
	executeCommand("rm", "-rf", ".git")

	coloredSource := colored("rm -rf .git", color.FgRed)
	coloredContent := fmt.Sprintf("\t-> %s", coloredSource)
	text(coloredContent, color.FgGreen)
}

/*
  startEngine
  Displays a welcome message and current version once libraries are ready.
*/
func startEngine() {
	text(fmt.Sprintf("Starting engine (%s)", VERSION), color.FgYellow)
}

/*
  copyFile
  Copies a file from a source to a destination using standard library.
*/
func copyFile(source, destination string) (err error) {
	in, inErr := os.Open(source)
	out, outErr := os.Create(destination)

	if inErr != nil {
		return inErr
	}

	if outErr != nil {
		return outErr
	}

	defer in.Close()
	defer out.Close()

	if _, err = io.Copy(out, in); err != nil {
		return err
	}

	return out.Sync()
}

/*
  deleteFile
  Deletes a standard file using standard library
*/
func deleteFile(path string) (err error) {
	return os.Remove(path)
}

/*
  moveFile
  Moves a standard file from a source to a destination
  Using both copyFile and deleteFile functions.
*/
func moveFile(source, destination string) (err error) {
	if err = copyFile(source, destination); err != nil {
		return err
	}

	if err = deleteFile(source); err != nil {
		return err
	}

	return err
}

/*
  substituteFile
  Dispatches the path information to either substituteInFile
  Or substituteInPath depending of if a path is given or not
*/
func substituteFile(path, variable, value *string, indent *int) (err error) {
	if *path != "" {
		return substituteInFile(path, variable, value, indent)
	} else {
		return substituteInPath(variable, value, indent)
	}
}

/*
  substituteInFile
  Replaces a content in a file using standard library
*/
func substituteInFile(path, variable, value *string, indent *int) (err error) {
	read, err := ioutil.ReadFile(*path)
	if err != nil {
		return err
	}

	indentedBlock := strings.Join(indentBlock(value, indent), "\n")
	newBytes := strings.Replace(string(read), *variable, indentedBlock, -1)

	return ioutil.WriteFile(*path, []byte(newBytes), 0)
}

func indentBlock(content *string, indent *int) (newData []string) {
	return Map(strings.Split(*content, "\n"), func(s string, i int) string {
		if i != 0 && indent != nil {
			return strings.Repeat(" ", *indent) + s
		} else {
			return s
		}
	})
}

/*
  substituteInPath
  Replaces a content if found in all files in the current directory
  This is recursive and can take a while for very large directories
*/
func substituteInPath(variable, value *string, indent *int) (err error) {
	err = filepath.Walk(".", func(filePath string, f os.FileInfo, err error) error {
		if !f.IsDir() {
			// text(fmt.Sprintf("\t-> Visiting: %s", filePath), color.FgWhite)
			if err = substituteInFile(&filePath, variable, value, indent); err != nil {
				return err
			}
		}

		return err
	})

	return err
}

/*
	processWarnings
*/
func processWarnings() {
	text("Parsing repository for valuable information", color.FgYellow)
	filepath.Walk(".", func(filePath string, f os.FileInfo, err error) error {
		if strings.Contains(filePath, ".cappuccino") {
			return nil
		}

		if !f.IsDir() {
			// text(fmt.Sprintf("\t-> %s", filePath), color.FgYellow, false)
			if err := processWarningInFile(&filePath); err != nil {

			}
		}

		return err
	})
}

func processWarningInFile(path *string) (err error) {
	f, err := os.Open(*path)
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(f)
	line := 1

	for scanner.Scan() {
		if bytes.Contains(scanner.Bytes(), []byte("[cappuccino-warning]")) {
			textContent := "\t-> Please make sure to setup needed information located %s in %s"
			content := fmt.Sprintf(textContent,
				colored(fmt.Sprintf("L-%03d", line), color.FgYellow),
				colored(*path, color.FgYellow))

			text(content, color.FgYellow)
		}

		line++
	}

	return scanner.Err()
}

/*
  Map
  Returns a new slice containing the results of applying the function f
  to each string in the original slice.
*/
func Map(vs []string, f func(string, int) string) []string {
	vsm := make([]string, len(vs))

	for i, v := range vs {
		vsm[i] = f(v, i)
	}

	return vsm
}

/*
  findRepoName
  Extracts Git repository name from a valid url
*/
func findRepoName(href string) string {
	re := regexp.MustCompile("/(.*).git")
	return re.FindStringSubmatch(href)[1]
}

/*
  displayVersion
  Displays the current Cappuccino version.
  Please refer to the CHANGELOG for related changes.
*/
func displayVersion(config *Config) {
	content := fmt.Sprintf("Detected version: %s", config.Version)
	text(content, color.FgYellow)
}

/*
  prefix
  Displays a prefix to all engine related messages
*/
func prefix() string {
	return fmt.Sprintf("Engine")
}

/*
  text
  Displays a message on the screen using a particular color
*/
func text(content string, attribute color.Attribute, returnOperator ...bool) {
	returnLine := true
	var printfContent string

	if len(returnOperator) > 0 {
		returnLine = returnOperator[0]
	}

	if returnLine {
		printfContent = "%s %s\n"
	} else {
		printfContent = "\r%s %s"
	}

	fmt.Printf(printfContent, colored(prefix(), attribute), content)
}

/*
  colored
  Displays a message on the screen using a particular color
*/
func colored(text string, attribute color.Attribute) string {
	return color.New(attribute).SprintFunc()(text)
}

/*
	getMessage
*/
func getMessage(index int) string {
	return "nothing"
}
