package legit_provenance_verifier

import (
	"flag"
	"fmt"
	"strings"

	intoto "github.com/in-toto/in-toto-golang/in_toto"
)

type ProvenanceChecks struct {
	SkipVerifySig bool
	RepoUrl       string
	Branch        string
	Tag           string
	BuilderId     string
	IsTagged      bool
}

type ProvenanceChecker interface {
	Check() error
}

type provenanceChecker struct {
	checks ProvenanceChecks
	ep     ExtendedProvenance
}

func newProvenanceChecker(statement *intoto.ProvenanceStatement, checks ProvenanceChecks) ProvenanceChecker {
	return provenanceChecker{
		checks: checks,
		ep:     ExtendedProvenance{Provenance: statement.Predicate},
	}
}

const (
	defaultBuilderID = `https://github.com/legit-labs/legit-provenance-action/.github/workflows/generate_provenance.yml`
	LegitBuildType   = `legit-security-provenance-generator`
	branchRefPrefix  = "refs/heads/"
	tagRefPrefix     = "refs/tags/"
)

func (pc *ProvenanceChecks) Flags() {
	flag.BoolVar(&pc.SkipVerifySig, "skip-signature-verification", false, "Skip signature verification (default: verify)")
	flag.StringVar(&pc.RepoUrl, "repo-url", "", "The source repository url (default: no check)")
	flag.StringVar(&pc.Branch, "branch", "", "The source branch (default: no check)")
	flag.StringVar(&pc.Tag, "tag", "", "The tag of the commit (default: no check)")
	flag.BoolVar(&pc.IsTagged, "is-tagged", false, "The commit is tagged (default: no check)")
	flag.StringVar(&pc.BuilderId, "builder-id", defaultBuilderID, fmt.Sprintf("The builder ID of the provenance generator (default: %v)", defaultBuilderID))
}

func (pc provenanceChecker) Check() error {
	if err := pc.verifyRepo(); err != nil {
		return err
	}

	if err := pc.verifyBuilderID(); err != nil {
		return err
	}

	if err := pc.verifyBuildType(); err != nil {
		return err
	}

	if err := pc.verifyBranch(); err != nil {
		return err
	}

	if err := pc.verifyIsTagged(); err != nil {
		return err
	}

	if err := pc.verifyTag(); err != nil {
		return err
	}

	return nil
}

func (pc provenanceChecker) verifyIsTagged() error {
	if !pc.checks.IsTagged {
		return nil
	}

	tagged, err := pc.ep.isTagged()
	if err != nil {
		return err
	}

	if !tagged {
		return fmt.Errorf("expected a tagged commit but the commit is untagged")
	}

	return nil
}

func (pc provenanceChecker) verifyTag() error {
	if pc.checks.Tag == "" {
		return nil
	}

	tag, err := pc.ep.Tag()
	if err != nil {
		return err
	}

	if tag != pc.checks.Tag {
		return fmt.Errorf("expected tag %v does not match actual: %v", pc.checks.Tag, tag)
	}

	return nil
}

func (pc provenanceChecker) verifyBranch() error {
	if pc.checks.Branch == "" {
		return nil
	}

	branch, err := pc.ep.Branch()
	if err != nil {
		return err
	}

	if branch != pc.checks.Branch {
		return fmt.Errorf("expected branch %v does not match actual: %v", pc.checks.Branch, branch)
	}

	return nil
}

func (pc provenanceChecker) verifyBuilderID() error {
	if pc.checks.BuilderId == "" {
		return nil
	}

	builderID := pc.ep.BuilderID()
	if !strings.HasPrefix(strings.ToLower(builderID), pc.checks.BuilderId) {
		return fmt.Errorf("expected builder ID %v does not match actual: %v", pc.checks.BuilderId, builderID)
	}

	return nil
}

func (pc provenanceChecker) verifyBuildType() error {
	buildType := pc.ep.BuildType()
	if buildType != LegitBuildType {
		return fmt.Errorf("expected build type %v does not match actual: %v", LegitBuildType, buildType)
	}

	return nil
}

func (pc provenanceChecker) verifyRepo() error {
	if pc.checks.RepoUrl == "" {
		return nil
	}

	repo, err := pc.ep.ProvenanceRepoUrl()
	if err != nil {
		return err
	}

	if repo != pc.checks.RepoUrl {
		return fmt.Errorf("expected repo %v does not match actual: %v", pc.checks.RepoUrl, repo)
	}

	return nil
}
