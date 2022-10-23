package legit_provenance_verifier

import (
	"fmt"
	"strings"

	slsa "github.com/in-toto/in-toto-golang/in_toto/slsa_provenance/v0.2"
)

type ExtendedProvenance struct {
	Provenance slsa.ProvenancePredicate
}

func (ep ExtendedProvenance) InvocationEnv() (map[string]interface{}, error) {
	env, ok := ep.Provenance.Invocation.Environment.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("failed to pull environment info")
	}

	return env, nil
}

func (ep ExtendedProvenance) isTagged() (bool, error) {
	env, err := ep.InvocationEnv()
	if err != nil {
		return false, err
	}

	reftype, ok := env["github_ref_type"]
	if !ok {
		return false, fmt.Errorf("failed to check github ref type")
	}

	isTag := reftype == "tag"

	return isTag, nil
}

func (ep ExtendedProvenance) Tag() (string, error) {
	env, err := ep.InvocationEnv()
	if err != nil {
		return "", err
	}

	tagged, err := ep.isTagged()
	if err != nil {
		return "", err
	}

	if !tagged {
		return "", fmt.Errorf("trying to check tag for an untagged version")
	}

	tag, ok := env["github_ref"]
	if !ok {
		return "", fmt.Errorf("failed to pull ref (tag)")
	}

	asString, ok := tag.(string)
	if !ok {
		return "", fmt.Errorf("unexpected type of tag: %T", tag)
	}

	clean := strings.TrimPrefix(asString, tagRefPrefix)

	return clean, nil
}

func (ep ExtendedProvenance) GithubEventPayload() (map[string]interface{}, error) {
	env, err := ep.InvocationEnv()
	if err != nil {
		return nil, err
	}

	event, ok := env["github_event_payload"]
	if !ok {
		return nil, fmt.Errorf("failed to pull github event payload")
	}

	asMap, ok := event.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected type of github event payload: %T", event)
	}

	return asMap, nil
}

func (ep ExtendedProvenance) Branch() (string, error) {
	tagged, err := ep.isTagged()
	if err != nil {
		return "", err
	}

	var b interface{}
	var ok bool
	if tagged {
		repo, err := ep.GithubEventPayload()
		if err != nil {
			return "", err
		}

		b, ok = repo["base_ref"]
	} else {
		env, err := ep.InvocationEnv()
		if err != nil {
			return "", err
		}

		b, ok = env["github_ref"]
	}

	if !ok {
		return "", fmt.Errorf("failed to pull base ref (branch)")
	}

	branch, ok := b.(string)
	if !ok {
		return "", fmt.Errorf("unexpected branch type: %T\n", b)
	}

	clean := strings.TrimPrefix(branch, branchRefPrefix)

	return clean, nil
}

func (ep ExtendedProvenance) BuilderID() string {
	return ep.Provenance.Builder.ID
}

func (ep ExtendedProvenance) BuildType() string {
	return ep.Provenance.BuildType
}

func (ep ExtendedProvenance) ProvenanceRepoInfo() (map[string]interface{}, error) {
	event, err := ep.GithubEventPayload()
	if err != nil {
		return nil, err
	}

	repo, ok := event["repository"]
	if !ok {
		return nil, fmt.Errorf("failed to pull repository info")
	}

	asMap, ok := repo.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected type of repository info: %T", repo)
	}

	return asMap, nil
}

func (ep ExtendedProvenance) ProvenanceRepoUrl() (string, error) {
	repo, err := ep.ProvenanceRepoInfo()
	if err != nil {
		return "", err
	}

	url, ok := repo["html_url"]
	if !ok {
		return "", fmt.Errorf("failed to pull repository url")
	}

	asString, ok := url.(string)
	if !ok {
		return "", fmt.Errorf("unexpected type of repository url: %T", url)
	}

	return asString, nil
}
