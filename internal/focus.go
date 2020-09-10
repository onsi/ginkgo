package internal

import (
	"regexp"
	"strings"

	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/types"
)

/*
	If a container marked as focus has a descendant that is also marked as focus, Ginkgo's policy is to
	unmark the container's focus.  This gives developers a more intuitive experience when debugging specs.
	It is common to focus a container to just run a subset of specs, then identify the specific specs within the container to focus -
	this policy allows the developer to simply focus those specific specs and not need to go back and turn the focus off of the container:

	As a common example, consider:

		FDescribe("something to debug", function() {
			It("works", function() {...})
			It("works", function() {...})
			FIt("doesn't work", function() {...})
			It("works", function() {...})
		})

	here the developer's intent is to focus in on the `"doesn't work"` spec and not to run the adjacent specs in the focused `"something to debug"` container.
	The nested policy applied by this function enables this behavior.
*/
func ApplyNestedFocusPolicyToTree(tree TreeNode) TreeNode {
	var walkTree func(tree TreeNode) (TreeNode, bool)
	walkTree = func(tree TreeNode) (TreeNode, bool) {
		if tree.Node.MarkedPending {
			return tree, false
		}
		hasFocusedDescendant := false
		processedTree := TreeNode{Node: tree.Node}
		for _, child := range tree.Children {
			processedChild, childHasFocus := walkTree(child)
			hasFocusedDescendant = hasFocusedDescendant || childHasFocus
			processedTree = AppendTreeNodeChild(processedTree, processedChild)
		}
		processedTree.Node.MarkedFocus = processedTree.Node.MarkedFocus && !hasFocusedDescendant
		return processedTree, processedTree.Node.MarkedFocus || hasFocusedDescendant
	}

	out, _ := walkTree(tree)
	return out
}

/*
	Ginkgo supports focussing specs using `FIt`, `FDescribe`, etc. - this is called "programmatic focus"
	It also supports focussing specs using regular expressions on the command line (`-focus=`, `-skip=`).
	The CLI regular expressions take precedence.

	This function sets the `Skip` property on specs by applying Ginkgo's focus policy:
	- If there are no CLI arguments and no programmatic focus, do nothing.
	- If there are no CLi arguments but a spec somewhere has programmatic focus, skip any specs that have no programmatic focus.
	- If there are CLI arguments parse them and skip any specs that either don't match the filter regexp or do match* the skip regexp.

	Lastly, `config.RegexScansFilePath` allows the regular exprressions to match against the spec's filepath as well as the spec's text.

	*Note:* specs with pending nodes are Skipped when created by NewSpec.
*/
func ApplyFocusToSpecs(specs Specs, description string, config config.GinkgoConfigType) (Specs, bool) {
	focusString := strings.Join(config.FocusStrings, "|")
	skipString := strings.Join(config.SkipStrings, "|")

	type SkipCheck func(spec Spec) bool

	// by default, skip any specs marked pending
	skipChecks := []SkipCheck{func(spec Spec) bool { return spec.Nodes.HasNodeMarkedPending() }}
	hasProgrammaticFocus := false

	if focusString == "" && skipString == "" {
		// check for programmatic focus
		for _, spec := range specs {
			if spec.Nodes.HasNodeMarkedFocus() && !spec.Nodes.HasNodeMarkedPending() {
				skipChecks = append(skipChecks, func(spec Spec) bool { return !spec.Nodes.HasNodeMarkedFocus() })
				hasProgrammaticFocus = true
				break
			}
		}
	}

	//the text to match when applying regexp filtering
	textToMatch := func(spec Spec) string {
		textToMatch := description + " " + spec.Text()
		if config.RegexScansFilePath {
			textToMatch += " " + spec.FirstNodeWithType(types.NodeTypeIt).CodeLocation.FileName
		}
		return textToMatch
	}

	if focusString != "" {
		// skip specs that don't match the focus string
		re := regexp.MustCompile(focusString)
		skipChecks = append(skipChecks, func(spec Spec) bool { return !re.MatchString(textToMatch(spec)) })
	}

	if skipString != "" {
		// skip specs that match the skip string
		re := regexp.MustCompile(skipString)
		skipChecks = append(skipChecks, func(spec Spec) bool { return re.MatchString(textToMatch(spec)) })
	}

	// skip specs if shouldSkip() is true.  note that we do nothing if shouldSkip() is false to avoid overwriting skip status established by the node's pending status
	processedSpecs := Specs{}
	for _, spec := range specs {
		for _, skipCheck := range skipChecks {
			if skipCheck(spec) {
				spec.Skip = true
				break
			}
		}
		processedSpecs = append(processedSpecs, spec)
	}

	return processedSpecs, hasProgrammaticFocus
}
