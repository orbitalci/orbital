extern crate structopt;
use structopt::StructOpt;

extern crate clap;
use structopt::clap::arg_enum;

// Copied definition from https://docs.rs/clap/2.33.0/clap/enum.Shell.html

arg_enum! {
  #[derive(Debug, StructOpt)]
  #[structopt(rename_all = "kebab-case")]
  pub enum Shell {
      Bash,
      Fish,
      Zsh,
      PowerShell,
      Elvish,
  }
}

impl From<clap::Shell> for SubcommandOption {
    fn from(s: clap::Shell) -> Self {
        match s {
            clap::Shell::Bash => SubcommandOption { shell: Shell::Bash },
            clap::Shell::Fish => SubcommandOption { shell: Shell::Fish },
            clap::Shell::Zsh => SubcommandOption { shell: Shell::Zsh },
            clap::Shell::PowerShell => SubcommandOption {
                shell: Shell::PowerShell,
            },
            clap::Shell::Elvish => SubcommandOption {
                shell: Shell::Elvish,
            },
        }
    }
}

impl From<SubcommandOption> for clap::Shell {
    fn from(s: SubcommandOption) -> Self {
        match s {
            SubcommandOption { shell: Shell::Bash } => clap::Shell::Bash,
            SubcommandOption { shell: Shell::Fish } => clap::Shell::Fish,
            SubcommandOption { shell: Shell::Zsh } => clap::Shell::Zsh,
            SubcommandOption {
                shell: Shell::PowerShell,
            } => clap::Shell::PowerShell,
            SubcommandOption {
                shell: Shell::Elvish,
            } => clap::Shell::Elvish,
        }
    }
}

#[derive(Debug, StructOpt)]
#[structopt(rename_all = "kebab_case")]
pub struct SubcommandOption {
    #[structopt(possible_values = &Shell::variants(), case_insensitive = true)]
    shell: Shell,
}
