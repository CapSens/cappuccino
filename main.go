package main

import (
  "os"
  "fmt"
  "log"
  "bytes"
  "regexp"
  "strings"
  "os/exec"
  "io/ioutil"
  "gopkg.in/yaml.v2"
  "github.com/fatih/color"
  "github.com/jessevdk/go-flags"
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
  Name     string
  Commands []ActionCommand
}

/*
  ActionCommand
  Structure mirroring the format of a valid command if a config file.
  Consists of a type and a string command.
*/
type ActionCommand struct {
  Type    string
  Command string
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
    text(fmt.Sprintf("Git url format successfuly verified"), color.FgYellow)
  } else {
    text(fmt.Sprintf("Git url format is not valid"), color.FgRed)
    os.Exit(0)
  }
}

/*
  cloneRepo
  Clones a specific branch of git repository.
  Same logic could be applied for an SVN repo.
*/
func cloneRepo(href string, branch string) {
  text(fmt.Sprintf("Cloning git repository (branch: %s)", branch), color.FgYellow)
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

  err = yaml.Unmarshal(content, &config)
  if err != nil {
    log.Fatalf("Error: %v", err)
  }

  displayVersion(&config)
  parseConfig(&config)
}

/*
  parseConfig
  Takes a Config pointer in argument and loops through the list
  of actions and commands, executing one after another in a
  thread safe executeCommand function.
*/
func parseConfig(config *Config) {
  text("Starting actions execution", color.FgYellow)
  removeGitDirectory()

  for i := 0; i < len(config.Actions); i++ {
    action := config.Actions[i]
    text(action.Name, color.FgGreen)

    for j := 0; j < len(action.Commands); j++ {
      command := action.Commands[j]
      content := fmt.Sprintf("\t-> %s", command.Command)
      text(content, color.FgGreen)

      executableCommand := strings.Split(command.Command, " ")
      executeCommand(executableCommand[0], executableCommand[1:]...)
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

  err := cmd.Run()

  if err != nil {
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
  text("Detected version: " + config.Version, color.FgYellow)
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
  _color := color.New(attribute).SprintFunc()
  fmt.Printf("%s %s\n", _color(prefix()), content)
}
