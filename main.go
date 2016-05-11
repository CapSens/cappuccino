package main

import (
	"bytes"
	"fmt"
	"github.com/fatih/color"
	"github.com/jessevdk/go-flags"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

const VERSION = "0.0.1"

type GitOptions struct {
	GitUrl string `short:"g" long:"git" description:"Git repository to clone"`
	Branch string `short:"b" long:"branch" description:"Branch to work with" default:"master"`
}

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
	Content []ActionContent
}

/*
  ActionContent
  Structure mirroring the format of a valid command if a config file.
  Consists of a type and a string command.
*/
type ActionContent struct {
	Type        string
	Path        string
	Command     string
	Source      string
	Destination string
	Arguments   ActionCommandArgument
}

type ActionCommandArgument struct {
	Variable string
	Path     string
	Value    string
}

func main() {
	var opts Options
	var parser = flags.NewParser(&opts, flags.Default)

	if _, err := parser.Parse(); err != nil {
		os.Exit(1)
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
	config := Config{}
	repoName := "app"
	os.Chdir(repoName)
	content, err := ioutil.ReadFile(".cappuccino.yml")

	if err != nil {
		text("Error opening .cappuccino.yml file", color.FgRed)
		os.Exit(0)
	}

	text("File .cappuccino.yml detected", color.FgYellow)

	if err := yaml.Unmarshal(content, &config); err != nil {
		log.Fatalf("Error: %v", err)
	}

	displayVersion(&config)
	processConfig(&config)
}

/*
  processConfig
  Takes a Config pointer in argument and loops through the list
  of actions and commands, executing one after another in a
  thread safe executeCommand function.
*/
func processConfig(config *Config) {
	text("Starting actions execution", color.FgYellow)
	removeGitDirectory()

	for i := 0; i < len(config.Actions); i++ {
		processAction(&config.Actions[i])
	}
}

func processAction(action *Action) {
	text(action.Name, color.FgGreen)

	for j := 0; j < len(action.Content); j++ {
		processContent(&action.Content[j])
	}
}

func processContent(content *ActionContent) {
	if content.Type == "exec" {
		command := content.Command
		coloredContent := fmt.Sprintf("\t-> %s", command)
		text(coloredContent, color.FgGreen)

		executableCommand := strings.Split(command, " ")
		executeCommand(executableCommand[0], executableCommand[1:]...)
	}

	if content.Type == "replace" {
		variable := content.Arguments.Variable
		value := content.Arguments.Value
		path := content.Arguments.Path

		read, err := ioutil.ReadFile(path)
		if err != nil {
			text(err.Error(), color.FgRed)
			os.Exit(0)
		}

		varName := fmt.Sprintf("[cappuccino-var-%s]", variable)
		newContent := strings.Replace(string(read), varName, value, -1)

		if err := ioutil.WriteFile(path, []byte(newContent), 0); err != nil {
			text(err.Error(), color.FgRed)
			os.Exit(0)
		}

		coloredName := colored(variable, color.FgCyan)
		coloredContent := fmt.Sprintf("\t-> %s", coloredName)
		text(coloredContent, color.FgGreen)
	}

	if content.Type == "copy" {
		source := content.Source
		destination := content.Destination

		coloredSource := colored(source, color.FgMagenta)
		coloredContent := fmt.Sprintf("\t-> %s", coloredSource)
		text(coloredContent, color.FgGreen)

		executeCommand("cp", source, destination)
	}

	if content.Type == "move" {
		source := content.Source
		destination := content.Destination

		coloredSource := colored(source, color.FgMagenta)
		coloredContent := fmt.Sprintf("\t-> %s", coloredSource)
		text(coloredContent, color.FgGreen)

		executeCommand("mv", source, destination)
	}

	if content.Type == "delete" {
		path := content.Path

		coloredSource := colored(path, color.FgMagenta)
		coloredContent := fmt.Sprintf("\t-> %s", coloredSource)
		text(coloredContent, color.FgGreen)

		executeCommand("rm", path)
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

func removeGitDirectory() {
	text("Removing existing .git folder", color.FgGreen)
	executeCommand("rm", "-rf", ".git")
}

/*
  startEngine
  Displays a welcome message and current version once libraries are ready.
*/
func startEngine() {
	text(fmt.Sprintf("Starting engine (%s)", VERSION), color.FgYellow)
}

/*
  displayVersion
  Displays the current Cappuccino version.
  Please refer to the CHANGELOG for related changes.
*/
func displayVersion(config *Config) {
	text("Detected version: "+config.Version, color.FgYellow)
}

/*
  prefix
  Displays a prefix to all engine related messages
*/
func prefix() string {
	return fmt.Sprintf(strings.ToUpper("engine"))
}

/*
  text
  Displays a message on the screen using a particular color
*/
func text(content string, attribute color.Attribute) {
	fmt.Printf("%s %s\n", colored(prefix(), attribute), content)
}

func colored(text string, attribute color.Attribute) string {
	return color.New(attribute).SprintFunc()(text)
}
