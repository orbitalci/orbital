use git2;
use git_meta::git_info;

use std::env;
use std::process::Command;

#[test]
fn current_commit() -> Result<(), git2::Error> {
    let shell_current_commit = format!(
        "{:?}",
        String::from_utf8(
            Command::new("git")
                .args(&["rev-parse", "HEAD"])
                .output()
                .expect("Failed to get current commit from `git`")
                .stdout
        )
        .unwrap()
    );

    // We want to use the repo that we're currently in, which is currently rooted two-levels up...
    let current_dir_abs_path = format!(
        "{}/../..",
        env::current_dir()
            .expect("Could not get current directory")
            .display()
    );

    // We want to use the current active branch
    let branch = None;
    // We want to use the HEAD commit
    let commit = None;

    // We're just gonna match the shell output. Wrap with Raw quotes, and raw newline
    let lib_current_commit = format!(
        "\"{}\\n\"",
        git_info::get_git_info_from_path(&current_dir_abs_path, &branch, &commit)?.id
    );

    assert_eq!(shell_current_commit, lib_current_commit);
    Ok(())
}
