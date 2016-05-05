package command

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	git "gopkg.in/src-d/go-git.v3"
)

// GetCommand is a Command implementation download and build template from remote location.
type GetCommand struct {
	Meta
}

func (c *GetCommand) Help() string {
	helpText := `
Usage: packer get [options] repository

  Get remote repository to local dir at specified revision if it present.

Options:

  -d=path                Destination dir to put files.
  -r=https://xxx.xx      Remote location to fetch
  -k=false               Keep destination dir after build
  -f=false               Only fetch template
`

	return strings.TrimSpace(helpText)
}

func (c *GetCommand) Run(args []string) int {
	var dest, remote string
	var keep, fetch bool
	var err error

	flags := c.Meta.FlagSet("get", FlagSetVars)
	flags.Usage = func() { c.Ui.Error(c.Help()) }
	flags.StringVar(&dest, "d", "", "d")
	flags.StringVar(&remote, "r", "", "r")
	flags.BoolVar(&fetch, "f", false, "f")
	flags.BoolVar(&keep, "k", false, "k")
	if err := flags.Parse(args); err != nil {
		return 1
	}

	args = flags.Args()
	if len(args) != 1 {
		flags.Usage()
		return 1
	}

	if dest == "" {
		dest, err = ioutil.TempDir("", "packer-get-")
		if err != nil {
			fmt.Printf("err: %s\n", err.Error())
			return 2
		}
	} else {
		if _, err = os.Stat(dest); err == nil {
			fmt.Printf("err: destination dir must not exists")
			return 2
		}
		if err = os.MkdirAll(dest, os.FileMode(0755)); err != nil {
			fmt.Printf("err: %s\n", err.Error())
			return 2
		}
	}

	u, err := url.Parse(remote)
	if err != nil {
		fmt.Printf("err: %s\n", err.Error())
		return 2
	}
	switch u.Scheme {
	default:
		fmt.Printf("scheme %q not supported", u.Scheme)
	case "git", "git+http", "git+https":
		if strings.HasPrefix(remote, "git+") {
			remote = remote[4:]
		}
		if err = getGit(remote, dest); err != nil {
			fmt.Printf("err: %s\n", err.Error())
			return 2
		}
	}

	if !keep {
		defer os.RemoveAll(dest)
	}

	if !fetch {
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Printf("err: %s\n", err.Error())
			return 2
		}
		err = os.Chdir(dest)
		if err != nil {
			fmt.Printf("err: %s\n", err.Error())
			return 2
		}
		defer os.Chdir(cwd)
		b := &BuildCommand{c.Meta}
		return b.Run(args)
	}
	return 0
}

func (c *GetCommand) Synopsis() string {
	return "Get template from remote location and built it"
}

func getGit(src string, dst string) error {
	repo, err := git.NewRepository(src, nil)
	if err != nil {
		return err
	}

	if err := repo.Pull(git.DefaultRemoteName, "refs/heads/master"); err != nil {
		return err
	}

	hash, err := repo.Remotes[git.DefaultRemoteName].Head()
	if err != nil {
		return err
	}

	commit, err := repo.Commit(hash)
	if err != nil {
		return err
	}

	fiter := commit.Tree().Files()
	defer fiter.Close()

	for {
		f, err := fiter.Next()
		if err != nil {
			if err == io.EOF && f == nil {
				break
			}
			return err
		}
		path := filepath.Dir(filepath.Join(dst, f.Name))
		if err = os.MkdirAll(path, os.FileMode(0755)); err != nil {
			return err
		}
		r, err := f.Reader()
		if err != nil {
			return err
		}
		if f.Mode == 40960 {
			buf, err := ioutil.ReadAll(r)
			if err != nil {
				return err
			}
			dst := string(buf)
			err = os.Symlink(dst, filepath.Join(path, filepath.Base(f.Name)))
			if err != nil {
				return err
			}
		} else {
			fp, err := os.OpenFile(filepath.Join(path, filepath.Base(f.Name)), os.O_WRONLY|os.O_CREATE|os.O_EXCL, f.Mode)
			if err != nil {
				return err
			}
			_, err = io.Copy(fp, r)
			if err != nil {
				r.Close()
				fp.Close()
				return err
			}
			fp.Close()
		}
		r.Close()
	}
	return nil
}
