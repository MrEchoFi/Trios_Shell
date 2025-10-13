

// A custom “Trios terminal” shell for cybersecurity professionals.

//
package main

import (
 "bufio"
 "fmt"
 //"io"
 "os"
 "os/exec"
 "os/signal"
 "strings"
 "syscall"
  "github.com/peterh/liner"
)

var torProcess *os.Process = nil

func main() {
 printLogo()
 signalhandler()

    line := liner.NewLiner()
    defer line.Close()
    line.SetCtrlCAborts(true) // Allow Ctrl+C to abort input

 for {
  // Display prompt
  //fmt.Print("Trios> ")
 cwd, _ := os.Getwd()
  home, _ := os.UserHomeDir()
  displayCwd := cwd
  if strings.HasPrefix(cwd, home) {
    displayCwd = "~" + cwd[len(home):]
  }
  // Bluish color: \033[34m (bright blue), reset: \033[0m
 prompt := fmt.Sprintf("\033[1;34m[~]Trios~%s>\033[0m ", displayCwd)
 fmt.Print(prompt)
 input, err := line.Prompt("")
        if err == liner.ErrPromptAborted {
            fmt.Println("^C")
            continue
        }
        if err != nil {
            fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
            break
        }
        input = strings.TrimSpace(input)
        if input == "" {
            continue
        }
        line.AppendHistory(input)

  // Handle built-in “exit”
  if input == "exit" || input == "quit" {
   break
  }

  // Parse first word to decide if it’s a custom command
  fields := strings.Fields(input)
  cmdName := fields[0]

  switch cmdName {
  case "triostor":
   // triostor start | stop | status
   handleTriosTor(fields[1:])

  case "sandbox":
   // sandbox <command...>
   if len(fields) < 2 {
    fmt.Println("Usage: sandbox <command...>")
    continue
   }
   Sandbox(fields[1:])

  case "scan":
   // scan <path...>
   if len(fields) < 2 {
    fmt.Println("Usage: scan <path...>")
    continue
   }
   handleScan(fields[1:])

  case "secureinput":
   // Secure input stub: hides echo while typing
   handleSecureInput()

   if len(fields) < 2 {
      fmt.Println("Usage: sandbox <command...>")
      continue
    }
    Sandbox(fields[1:])

	case "cd":
    if len(fields) < 2 {
      fmt.Println("Usage: cd <directory>")
      continue
    }
    if err := os.Chdir(fields[1]); err != nil {
      fmt.Fprintf(os.Stderr, "cd error: %v\n", err)
    }
    continue

  case "help":
   printTriosHelp() // <-- trios help cmd added

  case "?":
 questionmark()

  default:
   // All other commands get passed directly to Bash
   runbash(input)
  }
 }

 fmt.Println("~ Goodbye from Trios ~")
 //fmt.Printf("\033[1;34mTrios~%s>\033[0m ", displayCwd)

}

// printLogo prints a simple ASCII art banner for Trios Terminal.
func printLogo() {
 logo := `
   
       
  XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX  
 XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX 
XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX 
 XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX  
  XXXXXXXXXX                 XXXXXXXXXXXXXXXXXXXXXX               XXXXXXXXXX   
   XXXXXXXXXX                XXXXXXXXXXXXXXXXXXXXXX              XXXXXXXXXX    
    XXXXXXXXXXXXXX           XXXXXXXXXXXXXXXXXXXXXX         XXXXXXXXXXXXXX     
     XXXXXXXXXXXX                  XXXXXXXXXXX               XXXXXXXXXXXX      
      XXXXXXXXXX                   XXXXXXXXXXX                XXXXXXXXXX       
       XXXXXXXX                    XXXXXXXXXXX                 XXXXXXXX        
        XXXXXX                     XXXXXXXXXXX                  XXXXXX         
         XXXX                      XXXXXXXXXXX                   XXXX          
          XX                       XXXXXXXXXXX                    XX           
                                   XXXXXXXXXXX                                 
                                   XXXXXXXXXXX                                 
                                   XXXXXXXXXXX                                 
                                   XXXXXXXXXXX                                 
                                   XXXXXXXXXXX                                 
                                   XXXXXXXXXXX                                 
                                   XXXXXXXXXXX                                 
                                   XXXXXXXXXXX                                 
                                   XXXXXXXXXXX                                 
                                   XXXXXXXXXXX                                 
                                   XXXXXXXXXXX                                 
                                   XXXXXXXXXXX                                 
                                   XXXXXXXXXXX                                 
                                   XXXXXXXXXXX                                 
                                   XXXXXXXXXXX                                 
                                   XXXXXXXXXXX                                 
                                   XXXXXXXXXXX                                 
                                     XXXXXXX                                   
                                      XXXXX                                    
                                       XXX                                     
                                        X                                      
               

                                      Trios Terminal
                     (A Cybersecurity‐Focused Shell_Ebwer Community)
                     
`
 fmt.Println("\033[1;36m" + logo + "\033[0m")
}

// setupSignalHandler catches SIGINT (Ctrl+C) and ignores it, so the shell
// does not exit on Ctrl+C. Child processes (Bash, tor, etc.) still see Ctrl+C.
func signalhandler() {
 c := make(chan os.Signal, 1)
 signal.Notify(c, syscall.SIGINT)
 go func() {
  for {
   <-c
   // simply ignore and re-print prompt on next loop
  }
 }()
}

// runBash executes whatever the user typed by passing it to /bin/bash -c "<input>".
// Standard input, output, and error are directly tied to the terminal.

func runbash(input string) {
 cmd := exec.Command("/bin/bash", "-c", input)
 cmd.Stdin = os.Stdin
 cmd.Stdout = os.Stdout
 cmd.Stderr = os.Stderr

 if err := cmd.Run(); err != nil {
  fmt.Fprintf(os.Stderr, "Error: %v\n", err)
 }
}

// handleTriosTor manages “triostor start|stop|status” commands.
func handleTriosTor(args []string) {
 if len(args) == 0 {
  fmt.Println("Usage: triostor [start|stop|status]")
  return
 }

 switch args[0] {
 case "start":
  if torProcess != nil {
   fmt.Println("» Tor is already running.")
   return
  }
  fmt.Println("» Starting Tor…")
  cmd := exec.Command("tor")
  // Allow Tor to print to this terminal's stdout/stderr
  cmd.Stdout = os.Stdout
  cmd.Stderr = os.Stderr
  if err := cmd.Start(); err != nil {
   fmt.Fprintf(os.Stderr, "Failed to start Tor: %v\n", err)
   return
  }
  torProcess = cmd.Process
  fmt.Printf("» Tor started (PID %d).  Configure your apps to use SOCKS5 127.0.0.1:9050\n", torProcess.Pid)

 case "stop":
  if torProcess == nil {
   fmt.Println("» Tor is not running.")
   return
  }
  fmt.Printf("» Stopping Tor (PID %d)…\n", torProcess.Pid)
  if err := torProcess.Kill(); err != nil {
   fmt.Fprintf(os.Stderr, "Failed to kill Tor: %v\n", err)
  } else {
   fmt.Println("» Tor stopped.")
  }
  torProcess = nil

 case "status":
  if torProcess == nil {
   fmt.Println("» Tor is not running.")
  } else {
   fmt.Printf("» Tor is running (PID %d).\n", torProcess.Pid)
  }

 default:
  fmt.Println("Usage: triostor [start|stop|status]")
 }
}


// handleSandbox runs the given command inside a Docker container named kali:latest
// mounting the current directory to /workspace in the container. Adjust the image name as needed.
func Sandbox(args []string) {
 // Example: sandbox nmap -A 10.0.0.1
 // becomes: docker run --rm -it -v $(pwd):/workspace -w /workspace kali:latest bash -c "nmap -A 10.0.0.1"
 pwd, err := os.Getwd()
 if err != nil {
  fmt.Fprintf(os.Stderr, "Cannot get current directory: %v\n", err)
  return
 }

 // Reconstruct the inner command
 innerCmd := strings.Join(args, " ")
 dockerArgs := []string{
  "run", "--rm", "-it",
  "-v", pwd + ":/workspace",
  "-w", "/workspace",
  "kali:latest",          // ← make sure you have a kali:latest image, or change it to one you prefer
  "bash", "-c", innerCmd,
 }

 fmt.Printf("» Launching sandbox: docker %s\n", strings.Join(dockerArgs, " "))
 cmd := exec.Command("docker", dockerArgs...)
 cmd.Stdin = os.Stdin
 cmd.Stdout = os.Stdout
 cmd.Stderr = os.Stderr

 if err := cmd.Run(); err != nil {
  fmt.Fprintf(os.Stderr, "Sandbox error: %v\n", err)
 }
}

// handleScan invokes ClamAV’s clamscan -r <paths…>. It requires clamscan to be installed.
func handleScan(paths []string) {
 args := append([]string{"-r"}, paths...)
 cmd := exec.Command("clamscan", args...)
 cmd.Stdout = os.Stdout
 cmd.Stderr = os.Stderr

 fmt.Printf("» Running clamscan -r %s\n", strings.Join(paths, " "))
 if err := cmd.Run(); err != nil {
  fmt.Fprintf(os.Stderr, "clamscan error (non-zero exit is normal if viruses found): %v\n", err)
 }
}

// handleSecureInput toggles “stty -echo” so that the user’s typing is not echoed.
// This is a minimal stub to “hide” keystrokes (an anti-keylogger placeholder).
func handleSecureInput() {
 // Turn off echo
 cmdOff := exec.Command("stty", "-echo")
 cmdOff.Stdin = os.Stdin
 if err := cmdOff.Run(); err != nil {
  fmt.Fprintf(os.Stderr, "Error disabling echo: %v\n", err)
  return
 }

 fmt.Print("Password (input hidden): ")
 reader := bufio.NewReader(os.Stdin)
 _, _ = reader.ReadString('\n') // read entire line, but don’t show it

 // Turn echo back on
 cmdOn := exec.Command("stty", "echo")
 cmdOn.Stdin = os.Stdin
 if err := cmdOn.Run(); err != nil {
  fmt.Fprintf(os.Stderr, "Error re-enabling echo: %v\n", err)
  return
 }

 fmt.Println() // newline after password prompt
 fmt.Println("» Input captured securely (echo was off).")
}

func GitClone(args []string) {
  repoURL := args[0]
  fmt.Printf("» Cloning repository: %s\n", repoURL)
  cmd := exec.Command("git", "clone", repoURL)
  cmd.Stdout = os.Stdout
  cmd.Stderr = os.Stderr
  if err := cmd.Run(); err != nil {
    fmt.Fprintf(os.Stderr, "git clone error: %v\n", err)
  }
}

func printTriosHelp() {
 fmt.Println(`Trios Terminal - A Cybersecurity-Focused Shell

 Copyright (C) 2025 Ebwer Community_Md Abu Naser Nayeem

Type help' to see this list.
Type '?' to see the Info main menu (aka directory node).
Type help name' to find out more about the function name'.
Use info bash' to find out more about the shell in general.
Use man -k' or info' to find out more about commands not in this list.

Custom commands:
  triostor [start|stop|status]   Manage Tor anonymity daemon
  sandbox <command...>           Run a command inside a Kali Linux Docker sandbox
  scan <path...>                 Scan files or directories with ClamAV
  secureinput                    Enter input with echo disabled (anti-keylogger stub)
  help                           Show this help message
  exit, quit                     Exit the Trios Terminal
 job_spec [&]                    history [-c] [-d offset] [n]>
 (( expression ))                if COMMANDS; then COMMANDS; >
 . filename [arguments]          jobs [-lnprs] [jobspec ...] >
 :                               kill [-s sigspec | -n signum>
 [ arg... ]                      let arg [arg ...]
 [[ expression ]]                local [option] name[=value] >
 alias [-p] [name[=value] ... >  logout [n]
 bg [job_spec ...]               mapfile [-d delim] [-n count>
 bind [-lpsvPSVX] [-m keymap] >  popd [-n] [+N | -N]
 break [n]                       printf [-v var] format [argu>
 builtin [shell-builtin [arg .>  pushd [-n] [+N | -N | dir]
 caller [expr]                   pwd [-LP]
 case WORD in [PATTERN [| PATT>  read [-ers] [-a array] [-d d>
 cd [-L|[-P [-e]] [-@]] [dir]    readarray [-d delim] [-n cou>
 command [-pVv] command [arg .>  readonly [-aAf] [name[=value]
 export [-fn] [name[=value] ..>  typeset [-aAfFgiIlnrtux] nam>
 false                           ulimit [-SHabcdefiklmnpqrstu>
 fc [-e ename] [-lnr] [first] >  umask [-p] [-S] [mode]
 fg [job_spec]                   unalias [-a] name [name ...]>
 for NAME [in WORDS ... ] ; do>  unset [-f] [-v] [-n] [name .>
 for (( exp1; exp2; exp3 )); d>  until COMMANDS; do COMMANDS->
 function name { COMMANDS ; } >  variables - Names and meanin>
 getopts optstring name [arg .>  wait [-fn] [-p var] [id ...]>
 hash [-lr] [-p pathname] [-dt>  while COMMANDS; do COMMANDS->
 help [-dms] [pattern ...]       { COMMANDS ; }
 compgen [-abcdefgjksuv] [-o o>  return [n]
 complete [-abcdefgjksuv] [-pr>  select NAME [in WORDS ... ;]>
 compopt [-o|+o option] [-DEI]>  set [-abefhkmnptuvxBCEHPT] [>
 continue [n]                    shift [n]
 coproc [NAME] command [redire>  shopt [-pqsu] [-o] [optname >
 declare [-aAfFgiIlnrtux] [nam>  source filename [arguments]
 dirs [-clpv] [+N] [-N]          suspend [-f]
 disown [-h] [-ar] [jobspec ..>  test [expr]
 echo [-neE] [arg ...]           time [-p] pipeline
 enable [-a] [-dnps] [-f filen>  times
 eval [arg ...]                  trap [-lp] [[arg] signal_spe>
 exec [-cl] [-a name] [command>  true
 exit [n]                        type [-afptP] name [name ...>
 
`)
}

func questionmark() {
    fmt.Println(
`This is the Info main menu (aka directory node).
A few useful Info commands:

  'q' quits;
  'H' lists all Info commands;
  'h' starts the Info tutorial;
  'mTexinfo RET' visits the Texinfo manual, etc.

* Menu:

Basics
* Common options: (coreutils)Common options.
* Coreutils: (coreutils).       Core GNU (file, text, shell) utilities.
* Date input formats: (coreutils)Date input formats.
* Ed: (ed).                     The GNU line editor
* File permissions: (coreutils)File permissions.
                                Access modes.
* Finding files: (find).        Operating on files matching certain criteria.
* Time: (time).                 time

C++ libraries
* autosprintf: (autosprintf).   Support for printf format strings in C++.

Compression
* Gzip: (gzip).                 General (de)compression of files (lzw).
* Lzip: (lzip).                 LZMA lossless data compressor

Editors
* nano: (nano).                 Small and friendly text editor.

GNU Gettext Utilities
* autopoint: (gettext)autopoint Invocation.
                                Copy gettext infrastructure.
* envsubst: (gettext)envsubst Invocation.
                                Expand environment variables.
* gettextize: (gettext)gettextize Invocation.
                                Prepare a package for gettext.
* gettext: (gettext).           GNU gettext utilities.
* ISO3166: (gettext)Country Codes.
                                ISO 3166 country codes.
* ISO639: (gettext)Language Codes.
                                ISO 639 language codes.
* msgattrib: (gettext)msgattrib Invocation.
                                Select part of a PO file.
* msgcat: (gettext)msgcat Invocation.
                                Combine several PO files.
* msgcmp: (gettext)msgcmp Invocation.
                                Compare a PO file and template.
* msgcomm: (gettext)msgcomm Invocation.
                                Match two PO files.
* msgconv: (gettext)msgconv Invocation.
                                Convert PO file to encoding.
* msgen: (gettext)msgen Invocation.
                                Create an English PO file.
* msgexec: (gettext)msgexec Invocation.
                                Process a PO file.
* msgfilter: (gettext)msgfilter Invocation.
 Pipe a PO file through a filter.
* msgfmt: (gettext)msgfmt Invocation.
                                Make MO files out of PO files.
* msggrep: (gettext)msggrep Invocation.
                                Select part of a PO file.
* msginit: (gettext)msginit Invocation.
                                Create a fresh PO file.
* msgmerge: (gettext)msgmerge Invocation.
                                Update a PO file from template.
* msgunfmt: (gettext)msgunfmt Invocation.
                                Uncompile MO file into PO file.
* msguniq: (gettext)msguniq Invocation.
                                Unify duplicates for PO file.
* ngettext: (gettext)ngettext Invocation.
                                Translate a message with plural.
* xgettext: (gettext)xgettext Invocation.
                                Extract strings into a PO file.

GNU Utilities
* dirmngr-client: (gnupg).      X.509 CRL and OCSP client.
* dirmngr: (gnupg).             X.509 CRL and OCSP server.
* gpg-agent: (gnupg).           The secret key daemon.
* gpg2: (gnupg).                OpenPGP encryption and signing tool.
* gpgsm: (gnupg).               S/MIME encryption and signing tool.

Individual utilities
* arch: (coreutils)arch invocation.             Print machine hardware name.
* b2sum: (coreutils)b2sum invocation.           Print or check BLAKE2 digests.
* base32: (coreutils)base32 invocation.         Base32 encode/decode data.
* base64: (coreutils)base64 invocation.         Base64 encode/decode data.
* basename: (coreutils)basename invocation.     Strip directory and suffix.
* basenc: (coreutils)basenc invocation.         Encoding/decoding of data.
* bibtex: (web2c)bibtex invocation.             Maintaining bibliographies.
* cat: (coreutils)cat invocation.               Concatenate and write files.
* chcon: (coreutils)chcon invocation.           Change SELinux CTX of files.
* chgrp: (coreutils)chgrp invocation.           Change file groups.
* chmod: (coreutils)chmod invocation.           Change access permissions.
* chown: (coreutils)chown invocation.           Change file owners and groups.
* chroot: (coreutils)chroot invocation.         Specify the root directory.
* cksum: (coreutils)cksum invocation.           Print POSIX CRC checksum.
* cmp: (diffutils)Invoking cmp.                 Compare 2 files byte by byte.
* comm: (coreutils)comm invocation.             Compare sorted files by line.
* cp: (coreutils)cp invocation.                 Copy files.
* csplit: (coreutils)csplit invocation.         Split by context.
* cut: (coreutils)cut invocation.               Print selected parts of lines.
* date: (coreutils)date invocation.             Print/set system date and time.
* dd: (coreutils)dd invocation.                 Copy and convert a file.
* dircolors: (coreutils)dircolors invocation.   Color setup for ls.
* dirname: (coreutils)dirname invocation.       Strip last file name component.
* du: (coreutils)du invocation.                 Report file usage.
* dvicopy: (web2c)dvicopy invocation.           Virtual font expansion
* dvitomp: (web2c)dvitomp invocation.           DVI to MPX (MetaPost pictures).
* dvitype: (web2c)dvitype invocation.           DVI to human-readable text.
* echo: (coreutils)echo invocation.             Print a line of text.
* env: (coreutils)env invocation.               Modify the environment.
* expand: (coreutils)expand invocation.         Convert tabs to spaces.
* expr: (coreutils)expr invocation.             Evaluate expressions.
* factor: (coreutils)factor invocation.         Print prime factors
* false: (coreutils)false invocation.           Do nothing, unsuccessfully.
* find: (find)Invoking find.                    Finding and acting on files.
* fmt: (coreutils)fmt invocation.               Reformat paragraph text.
* fold: (coreutils)fold invocation.             Wrap long input lines.
* gftodvi: (web2c)gftodvi invocation.           Generic font proofsheets.
* gftopk: (web2c)gftopk invocation.             Generic to packed fonts.
* gftype: (web2c)gftype invocation.             GF to human-readable text.
* groups: (coreutils)groups invocation.         Print group names a user is in.
* gunzip: (gzip)Overview.                       Decompression.
* gzexe: (gzip)Overview.                        Compress executables.
* head: (coreutils)head invocation.             Output the first part of files.
* hostid: (coreutils)hostid invocation.         Print numeric host identifier.
* hostname: (coreutils)hostname invocation.     Print or set system name.
* id: (coreutils)id invocation.                 Print user identity.
* install: (coreutils)install invocation.       Copy files and set attributes.
* install-tl::                  Installing TeX Live.
* join: (coreutils)join invocation.             Join lines on a common field.
* kill: (coreutils)kill invocation.             Send a signal to processes.
* link: (coreutils)link invocation.             Make hard links between files.
* ln: (coreutils)ln invocation.                 Make links between files.
* locate: (find)Invoking locate.                Finding files in a database.
* logname: (coreutils)logname invocation.       Print current login name.
* ls: (coreutils)ls invocation.                 List directory contents.
* md5sum: (coreutils)md5sum invocation.         Print or check MD5 digests.
* mf: (web2c)mf invocation.                     Creating typeface families.
* mft: (web2c)mft invocation.                   Prettyprinting Metafont source.
* mkdir: (coreutils)mkdir invocation.           Create directories.
* mkfifo: (coreutils)mkfifo invocation.         Create FIFOs (named pipes).
* mknod: (coreutils)mknod invocation.           Create special files.
* mktemp: (coreutils)mktemp invocation.         Create temporary files.
* mltex: (web2c)MLTeX.                          Multi-lingual TeX.
* mpost: (web2c)mpost invocation.               Generating PostScript.
* mv: (coreutils)mv invocation.                 Rename files.
* nice: (coreutils)nice invocation.             Modify niceness.
* nl: (coreutils)nl invocation.                 Number lines and write files.
* nohup: (coreutils)nohup invocation.           Immunize to hangups.
* patch: (diffutils)Invoking patch.             Apply a patch to a file.
* patgen: (web2c)patgen invocation.             Creating hyphenation patterns.
* pathchk: (coreutils)pathchk invocation.       Check file name portability.
* pktogf: (web2c)pktogf invocation.             Packed to generic fonts.
* pktype: (web2c)pktype invocation.             PK to human-readable text.
* pltotf: (web2c)pltotf invocation.             Property list to TFM.
* pooltype: (web2c)pooltype invocation.         Display WEB pool files.
* pr: (coreutils)pr invocation.                 Paginate or columnate files.
* printenv: (coreutils)printenv invocation.     Print environment variables.
* printf: (coreutils)printf invocation.         Format and print data.
* ptx: (coreutils)ptx invocation.               Produce permuted indexes.
* pwd: (coreutils)pwd invocation.               Print working directory.
* readlink: (coreutils)readlink invocation.     Print referent of a symlink.
* realpath: (coreutils)realpath invocation.     Print resolved file names.
* rm: (coreutils)rm invocation.                 Remove files.
* rmdir: (coreutils)rmdir invocation.           Remove empty directories.
* runcon: (coreutils)runcon invocation.         Run in specified SELinux CTX.
* sdiff: (diffutils)Invoking sdiff.             Merge 2 files side-by-side.
* seq: (coreutils)seq invocation.               Print numeric sequences
* sha1sum: (coreutils)sha1sum invocation.       Print or check SHA-1 digests.
* sha2: (coreutils)sha2 utilities.              Print or check SHA-2 digests.
* shred: (coreutils)shred invocation.           Remove files more securely.
* shuf: (coreutils)shuf invocation.             Shuffling text files.
* sleep: (coreutils)sleep invocation.           Delay for a specified time.
* sort: (coreutils)sort invocation.             Sort text files.
* split: (coreutils)split invocation.           Split into pieces.
* stat: (coreutils)stat invocation.             Report file(system) status.
* stdbuf: (coreutils)stdbuf invocation.         Modify stdio buffering.
* stty: (coreutils)stty invocation.             Print/change terminal settings.
* sum: (coreutils)sum invocation.               Print traditional checksum.
* sync: (coreutils)sync invocation.             Sync files to stable storage.
* tac: (coreutils)tac invocation.               Reverse files.
* tail: (coreutils)tail invocation.             Output the last part of files.
* tangle: (web2c)tangle invocation.             WEB to Pascal.
* tee: (coreutils)tee invocation.               Redirect to multiple files.
* test: (coreutils)test invocation.             File/string tests.
* tex: (web2c)tex invocation.                   Typesetting.
* tftopl: (web2c)tftopl invocation.             TFM -> property list.
* timeout: (coreutils)timeout invocation.       Run with time limit.
* tlmgr::                       Managing TeX Live.
* touch: (coreutils)touch invocation.           Change file timestamps.
* tr: (coreutils)tr invocation.                 Translate characters.
* true: (coreutils)true invocation.             Do nothing, successfully.
* truncate: (coreutils)truncate invocation.     Shrink/extend size of a file.
* tsort: (coreutils)tsort invocation.           Topological sort.
* tty: (coreutils)tty invocation.               Print terminal name.
* uname: (coreutils)uname invocation.           Print system information.
* unexpand: (coreutils)unexpand invocation.     Convert spaces to tabs.
* uniq: (coreutils)uniq invocation.             Uniquify files.
* unlink: (coreutils)unlink invocation.         Removal via unlink(2).
* updatedb: (find)Invoking updatedb.            Building the locate database.
* uptime: (coreutils)uptime invocation.         Print uptime and load.
* users: (coreutils)users invocation.           Print current user names.
* vdir: (coreutils)vdir invocation.             List directories verbosely.
* vftovp: (web2c)vftovp invocation.             Virtual font -> virtual pl.
* vptovf: (web2c)vptovf invocation.             Virtual pl -> virtual font.
* wc: (coreutils)wc invocation.                 Line, word, and byte counts.
* wdiff: (wdiff)wdiff invocation.               Word difference finder.
* weave: (web2c)weave invocation.               WEB to TeX.
* who: (coreutils)who invocation.               Print who is logged in.
* whoami: (coreutils)whoami invocation.         Print effective user ID.
* xargs: (find)Invoking xargs.                  Operating on many files.
* yes: (coreutils)yes invocation.               Print a string indefinitely.
* zcat: (gzip)Overview.                         Decompression to stdout.
* zdiff: (gzip)Overview.                        Compare compressed files.
* zforce: (gzip)Overview.                       Force .gz extension on files.
* zgrep: (gzip)Overview.                        Search compressed files.
* zmore: (gzip)Overview.                        Decompression output by pages.

Libraries
* RLuserman: (rluserman).       The GNU readline library User's Manual.

Math
* bc: (bc).                     An arbitrary precision calculator language.

Network applications
* Wget: (wget).                 Non-interactive network downloader.

TeX
* Kpathsea: (kpathsea).         File lookup along search paths.
* TLbuild: (tlbuild).           TeX Live configuration and development.
* Web2c: (web2c).               TeX, Metafont, and companion programs.
* afm2tfm: (dvips)Invoking afm2tfm.
                                Making Type 1 fonts available to TeX.
* DVI-to-PostScript: (dvips).   Translating TeX DVI files to PostScript.
* dvips: (dvips)Invoking Dvips. DVI-to-PostScript translator.
* kpsewhich: (kpathsea)Invoking kpsewhich.
                                TeX file searching.
* mktexfmt: (kpathsea)mktex scripts.
                                Format (fmt/base/mem) generation.
* mktexlsr: (kpathsea)Filename database.
                                Update ls-R.
* mktexmf: (kpathsea)mktex scripts.
                                MF source generation.
* mktexpk: (kpathsea)mktex scripts.
                                PK bitmap generation.
                                * mktextex: (kpathsea)mktex scripts.
                                TeX source generation.
* mktextfm: (kpathsea)mktex scripts.
                                TeX font metric generation.

Texinfo documentation system
* info stand-alone: (info-stnd).
                                Read Info documents without Emacs.

Text creation and manipulation
* Diffutils: (diffutils).       Comparing and merging files.
* Word differences: (wdiff).    GNU wdiff and diff related tools.
* grep: (grep).                 Print lines that match patterns.
* sed: (sed).                   Stream EDitor.`)
}
