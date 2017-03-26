package main

import (
    "fmt"
    "runtime"
    "strings"
    "os"
    "os/exec"
    "os/user"
	"io/ioutil"
    "path/filepath"
    "github.com/fsouza/go-dockerclient"
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

func showHelp() {
    fmt.Println(`DIN - Docker IN. Command to ease the multi-language development

Usage: din <command> <arguments>

Commands:
    help                 shows the help
    list                 list the possible languages
    clean                deletes all the containers
    clean [<lang>]       deletes the specific language container
    update [<lang>]      forces to reconstruct the container
    install all|<lang>   install the container for <lang>. 'all' will install every possible language
    <lang>               executes elm binary on the file
    <lang>/<cmd>         executes the command <cmd> inside a container for the language`)

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

    fmt.Println("Updated '%'", tag)
    return true
}

func dinCleanCommand(tags ...string) {
    availableTagsList := availableTags()
    if len(tags) == 0 && confirm("delete all images"){
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
            fmt.Println("Language %s not found.", tag)
        }
    }
}

func dinUpdateCommand(tags ...string) {
    if len(tags) == 0 && confirm("update all images"){
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
            fmt.Printf("ERROR.Language %s not a valid candidate.\n", tag)
        }
    }
}

func dinInstallCommand(tag string) bool {
    if tag == "all" {
        for _, lang := range availableTags() {
            installLanguage(lang)
        }
    } else if contains(availableTags(), tag) {
        installLanguage(tag)
    } else {
        fmt.Printf("ERROR.Language %s not a valid candidate.\n", tag)
        showCandidates()
        return false
    }
    return true
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

func main () {
    dockerClient, _ = docker.NewClient("unix:///var/run/docker.sock")

    if len(os.Args) == 1 {
        showHelp()
        return
    }

    cmd := os.Args[1]

    switch cmd {
        case "help":
            showHelp()
        case "list":
            showCandidates()
        case "clean":
            dinCleanCommand(os.Args[2:]...)
        case "update":
            dinUpdateCommand(os.Args[2:]...)
        case "install":
            dinInstallCommand(os.Args[2])
        default:
            params := strings.Split(cmd, "/")
            if(len(params) == 1) {
                executeDin(params[0], "")
            } else {
                executeDin(params[0], params[1])
            }
    }
}
