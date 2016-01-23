package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
)

type Payload struct {
	Username string
	RepoName string
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	baseResp := BaseResponse{
		CreateRepositoryUrl: fmt.Sprintf(GetProtocol(false) + r.Host + GetRepoCreateUrl()),
		UserRepositoriesUrl: fmt.Sprintf(GetProtocol(false) + r.Host + GetReposUrl()),
		UserRepositoryUrl:   fmt.Sprintf(GetProtocol(false) + r.Host + GetRepoUrl()),
		BranchesUrl:         fmt.Sprintf(GetProtocol(false) + r.Host + GetBranchesUrl()),
		BranchUrl:           fmt.Sprintf(GetProtocol(false) + r.Host + GetBranchUrl()),
	}

	WriteIndentedJson(w, baseResp, "", "  ")
}

func repoCreateHandler(w http.ResponseWriter, r *http.Request) {
	var resp CreateResponse
	resp.ResponseMessage = "Unknown error. Follow README"
	resp.CloneUrl = ""

	wd, _ := os.Getwd()

	defer func() {
		WriteIndentedJson(w, resp, "", "  ")
		if err := os.Chdir(wd); err != nil {
			log.Println(err)
		}
	}()

	var payload Payload
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&payload); err != nil {
		log.Println(err)
		return
	} else {
		if payload.Username == "" || payload.RepoName == "" {
			log.Println("Empty username or reponame")
			return
		}
	}

	usrPath := UserPath(payload.Username)
	bareRepo := FormatRepoName(payload.RepoName)
	url := FormCloneURL(r.Host, payload.Username, bareRepo)

	if _, err := os.Stat(RepoPath(payload.Username, payload.RepoName)); err == nil {
		resp.ResponseMessage = fmt.Sprintf("repository already exists for %s", payload.Username)
		resp.CloneUrl = url
		return
	}

	if err := os.MkdirAll(usrPath, 0775); err != nil {
		resp.ResponseMessage = "error while creating user"
		return
	}

	if err := os.Chdir(usrPath); err != nil {
		resp.ResponseMessage = "error while creating new repository"
		return
	}

	cmd := exec.Command(config.GitPath, "init", "--bare", bareRepo)

	if err := cmd.Start(); err == nil {
		resp.CloneUrl = url
		resp.ResponseMessage = "Repository created successfully"
	} else {
		resp.ResponseMessage = "error while creating new repository"
		return
	}
	if err := cmd.Wait(); err != nil {
		log.Println("Error while waiting:", err)
		return
	}
}

func repoIndexHandler(w http.ResponseWriter, r *http.Request) {
	userName, _ := GetParamValues(r)
	var errJson Error
	list, ok := FindAllDir(UserPath(userName))
	if !ok {
		errJson = Error{Message: "repository not found"}
		WriteIndentedJson(w, errJson, "", "  ")
		return
	}
	var repo Repository
	repos := make([]Repository, 0)

	for i := 0; i < len(list); i++ {
		repo = GetRepository(r.Host, userName, list[i].Name())
		repos = append(repos, repo)
	}
	WriteIndentedJson(w, repos, "", "  ")
}

func repoShowHandler(w http.ResponseWriter, r *http.Request) {
	var errJson Error
	userName, repoName := GetParamValues(r)
	if ok := IsExistingRepository(RepoPath(userName, repoName)); !ok {
		errJson = Error{Message: "repository not found"}
		WriteIndentedJson(w, errJson, "", "  ")
		return
	}
	repo := GetRepository(r.Host, userName, FormatRepoName(repoName))
	WriteIndentedJson(w, repo, "", "  ")
}

func branchIndexHandler(w http.ResponseWriter, r *http.Request) {

}

func branchShowHandler(w http.ResponseWriter, r *http.Request) {

}