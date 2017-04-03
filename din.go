package main

import (
    "fmt"
    "runtime"
    "errors"
    "strings"
    "os"
    "os/exec"
    "os/user"
	"io/ioutil"
    "path/filepath"
    "github.com/fsouza/go-dockerclient"
    "github.com/urfave/cli"
)

var dockerClient *docker.Client

func getScriptPath() string {
    _, filename, _, _ := runtime.Caller(1)
    dir, _ := filepath.Abs(filepath.Dir(filename))
    return dir
}

func showCandidates() {
    for _, tag := range availableTags() {
        fmt.Println(tag)
    }
}

func confirm(action string) bool {
    var confirm string
    fmt.Printf("I'm about to proceed and %s.\n", action)
    fmt.Print("Are you sure? (y/n) ")
    fmt.Scanf("%s", &confirm)
    if confirm == "y" || confirm == "Y" {
        return true
    }
    return false
}

func availableTags() []string {
    scriptPath := getScriptPath()
	files, err := ioutil.ReadDir(filepath.Join(scriptPath, "lang"))
	if err != nil {
        return []string{}
	}

    languages := make([]string, len(files))
    for i, file := range files {
        languages[i] = strings.Replace(file.Name(), ".docker", "", -1)
    }
    return languages
}

func contains(s []string, e string) bool {
    for _, a := range s {
        if strings.Compare(a, e) == 0 {
            return true
        }
    }
    return false
}

func runImage(tag string, cmd string) bool {
    currentUser, _ := user.Current()
    currentDir, _ := os.Getwd()

    dockerCmd := exec.Command(
        "docker", "run",
        "-it",
        "-e", fmt.Sprintf("DIN_ENV_PWD=\"%s\"", currentDir),
        "-e", fmt.Sprintf("DIN_ENV_UID=%s", currentUser.Uid),
        "-e", fmt.Sprintf("DIN_ENV_USER=%s", currentUser.Username),
        "-e", fmt.Sprintf("DIN_COMMAND=%s", cmd),
        "-v", fmt.Sprintf("%s:/home/%s", currentUser.HomeDir, currentUser.Username),
        fmt.Sprintf("din/%s", tag),
    )
    dockerCmd.Stdin = os.Stdin
    dockerCmd.Stdout = os.Stdout
    dockerCmd.Stderr = os.Stderr
    dockerCmd.Run()
    return true
}

func buildImage(tag string) bool {
    opts := docker.BuildImageOptions{
        Name: fmt.Sprintf("din/%s", tag),
        Dockerfile: fmt.Sprintf("lang/%s.docker", tag),
        ContextDir: getScriptPath(),
        OutputStream: os.Stdout,
    }
    err := dockerClient.BuildImage(opts)
    return err == nil
}

func buildBaseImage() bool {
    opts := docker.BuildImageOptions{
        Name: "din/base",
        Dockerfile: "Dockerfile",
        ContextDir: filepath.Join(getScriptPath(), "base"),
        OutputStream: os.Stdout,
    }
    err := dockerClient.BuildImage(opts)
    return err == nil
}

func checkImageExists(tag string) bool {
    _, err := dockerClient.InspectImage(fmt.Sprintf("din/%s", tag))
    return err == nil
}

func installLanguage(tag string) bool {
    if !checkImageExists("base") {
        fmt.Println("Creating base image....")
        if !buildBaseImage() {
            fmt.Println("Cannot create the 'din/base' image. Exiting")
            return false
        }
        fmt.Println("Base image created.")
    }

    fmt.Println("Installing language {}...", tag)
    if !checkImageExists(tag) {
        if !buildImage(tag) {
            fmt.Printf("Cannot create the 'din/%s' image.\n", tag)
            return false
        }
    }
    fmt.Printf("Installed '%s'\n", tag)
    return true
}

func updateLanguage(tag string) bool {
    if !checkImageExists(tag) {
        fmt.Println("The language %s has not been installed. Cancel update", tag)
        return false
    }

    fmt.Printf("Updating language %s...\n", tag)

    if !buildImage(tag) {
        fmt.Printf("Cannot update the 'din/%s' image.\n", tag)
        return false
    }

    fmt.Printf("Updated '%s'\n", tag)
    return true
}

func cleanCommand(c *cli.Context) error {
    availableTagsList := availableTags()
    tags := c.Args()
    if c.NArg() == 0 && confirm("delete all images") {
        for _, tag := range availableTagsList {
            if checkImageExists(tag) {
                tags = append(tags, tag)
            }
        }
    }

    for _, tag := range tags {
        if contains(availableTagsList, tag) {
            if !checkImageExists(tag) {
                fmt.Printf("There is nothing to clean for language %s.\n", tag)
            } else {
                fmt.Printf("Cleaning %s.\n", tag)
                opts := docker.RemoveImageOptions{Force: true}
                dockerClient.RemoveImageExtended(fmt.Sprintf("din/%s", tag), opts)
            }
        } else {
            return fmt.Errorf("Language %s not found", tag)
        }
    }
    return nil
}

func updateCommand(c *cli.Context) error {
    tags := c.Args()
    if c.NArg() == 0 && confirm("update all images") {
        for _, tag := range availableTags() {
            if checkImageExists(tag) {
                tags = append(tags, tag)
            }
        }
    }

    for _, tag := range tags {
        if contains(availableTags(), tag) {
            if !checkImageExists(tag) {
                fmt.Printf("Language %s has not been installed.\n", tag)
            } else {
                fmt.Printf("Updating %s.\n", tag)
                updateLanguage(tag)
            }
        } else {
            return fmt.Errorf("Language %s not a valid candidate", tag)
        }
    }
    return nil
}

func installCommand(c *cli.Context) error {
    if c.NArg() == 0 {
        return errors.New("Must exist the language")
    }
    for _, tag := range c.Args() {
        if tag == "all" {
            for _, lang := range availableTags() {
                installLanguage(lang)
            }
        } else if contains(availableTags(), tag) {
            installLanguage(tag)
        } else {
            return fmt.Errorf("Language %s not a valid candidate", tag)
        }
    }
    return nil
}

func executeDin(tag string, cmd string) bool {
    if cmd == "" {
        cmd = tag
    }

    if !contains(availableTags(), tag) {
        fmt.Printf("ERROR: Can't find language: %s\n", tag)
        showCandidates()
        return false
    }


    if !checkImageExists("base") {
        fmt.Println("Creating base image....")
        if !buildBaseImage() {
            fmt.Println("Cannot create the 'din/base' image. Exiting")
            return false
        }
        fmt.Println("Base image created.")
    }

    if !checkImageExists(tag) {
        if !buildImage(tag) {
            fmt.Printf("Cannot create the 'din/%s' image.\n", tag)
            return false
        }
    }

    return runImage(tag, cmd)
}

func showCandidatesCommand(c *cli.Context) error {
      showCandidates()
      return nil
}

func mainCommand(c *cli.Context) error {
     if c.NArg() == 0 {
         cli.ShowAppHelp(c)
         return nil
     }
     params := strings.Split(c.Args()[0], "/")
     if(len(params) == 1) {
         executeDin(params[0], "")
     } else {
         executeDin(params[0], params[1])
     }
     return nil
}

func main () {
    dockerClient, _ = docker.NewClient("unix:///var/run/docker.sock")
    app := cli.NewApp()
    app.Commands = []cli.Command{
      {
        Name:    "list",
        Usage:   "list the possible languages",
        Action:  showCandidatesCommand,
      },
      {
        Name:    "clean",
        Usage:   "deletes a language container",
        Action:  cleanCommand,
      },
      {
        Name:    "update",
        Usage:   "update a language container",
        Action:  updateCommand,
      },
      {
        Name:    "install",
        Usage:   "install the container for <lang>. 'all' will install every possible language",
        ArgsUsage: "all|<lang>",
        Action:  installCommand,
      },
    }
    app.HideVersion = true
    app.UsageText = "din <command> <arguments>"
    app.Name = "Docker IN"
    app.Usage = "Command to ease the multi-language development"
    app.Action = mainCommand

    err := app.Run(os.Args)
    if err != nil {
        fmt.Printf("ERROR: %s\n", err)
    }
}
