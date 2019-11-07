/*
	trigger contains the logic for determining whether or not stages and/or commits pushed to a given
	  repository should be executed.
	this includes:
	  - checking if a given branch matches a regex list of acceptable branches
      - a parser for the sentence-like conditions that can be used in a build stage
	  - support for 3 types of trigger conditions:
		- match filepaths changed
		- match text in commit messages
   		- match branch
*/
package runtime
