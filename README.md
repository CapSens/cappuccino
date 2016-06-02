## Cappuccino
Cappuccino is a [Go](https://golang.org) project that helps developers avoid repeating tasks by defining a structured `.cappuccino.yml` config file.

Robots should do the hard work, that's why a config file contains from few to dozens of actions requiring executing commands, renaming, searching and replacing, generating from templates and more. Cappuccino is written in Go and is thus executable on all plateforms, without a need for an `LLVM` or an interpreter.

### Installation
```
sudo curl -sSo /usr/bin/cappuccino https://raw.githubusercontent.com/CapSens/cappuccino/master/main && sudo chmod 777 /usr/bin/cappuccino
```

### Manual download
Every new Cappuccino version is released using Github Releases and the latest release download links are available here:
```
https://github.com/CapSens/cappuccino/releases/latest
```

Here are all available plateforms:
```
### Darwin (Apple Mac)

 * [cappuccino\_0.1.3\_darwin\_386.zip](cappuccino_0.1.3_darwin_386.zip)
 * [cappuccino\_0.1.3\_darwin\_amd64.zip](cappuccino_0.1.3_darwin_amd64.zip)

### FreeBSD

 * [cappuccino\_0.1.3\_freebsd\_386.zip](cappuccino_0.1.3_freebsd_386.zip)
 * [cappuccino\_0.1.3\_freebsd\_amd64.zip](cappuccino_0.1.3_freebsd_amd64.zip)
 * [cappuccino\_0.1.3\_freebsd\_arm.zip](cappuccino_0.1.3_freebsd_arm.zip)

### Linux

 * [cappuccino\_0.1.3\_amd64.deb](cappuccino_0.1.3_amd64.deb)
 * [cappuccino\_0.1.3\_armhf.deb](cappuccino_0.1.3_armhf.deb)
 * [cappuccino\_0.1.3\_i386.deb](cappuccino_0.1.3_i386.deb)
 * [cappuccino\_0.1.3\_linux\_386.tar.gz](cappuccino_0.1.3_linux_386.tar.gz)
 * [cappuccino\_0.1.3\_linux\_amd64.tar.gz](cappuccino_0.1.3_linux_amd64.tar.gz)
 * [cappuccino\_0.1.3\_linux\_arm.tar.gz](cappuccino_0.1.3_linux_arm.tar.gz)

### MS Windows

 * [cappuccino\_0.1.3\_windows\_386.zip](cappuccino_0.1.3_windows_386.zip)
 * [cappuccino\_0.1.3\_windows\_amd64.zip](cappuccino_0.1.3_windows_amd64.zip)

### NetBSD

 * [cappuccino\_0.1.3\_netbsd\_386.zip](cappuccino_0.1.3_netbsd_386.zip)
 * [cappuccino\_0.1.3\_netbsd\_amd64.zip](cappuccino_0.1.3_netbsd_amd64.zip)
 * [cappuccino\_0.1.3\_netbsd\_arm.zip](cappuccino_0.1.3_netbsd_arm.zip)

### OpenBSD

 * [cappuccino\_0.1.3\_openbsd\_386.zip](cappuccino_0.1.3_openbsd_386.zip)
 * [cappuccino\_0.1.3\_openbsd\_amd64.zip](cappuccino_0.1.3_openbsd_amd64.zip)

### Other files

 * [control.tar.gz](.goxc-temp/control.tar.gz)
 * [data.tar.gz](.goxc-temp/data.tar.gz)
 * [LICENSE.md](LICENSE.md)
 * [README.md](README.md)

### Plan 9

 * [cappuccino\_0.1.3\_plan9\_386.zip](cappuccino_0.1.3_plan9_386.zip)
```

### Example of use
Let's say we need to clone a [Ruby on Rails](http://rubyonrails.org/) git repository and apply the following changes :
- Rename the `database.yml.example` file into `database.yml`.
- Delete `Procfile` and `bower.json` files.
- Substitute user defined variables in both `.ruby-version` and `.ruby-gemset` files.
- Create the gemset using `RVM`.
- Bundle install, create, migrate and seed the database.

First, here is what the config file would look like :

```yaml
engine: cappuccino
actions:
  - name: Copying database config file
    type: copy
    content:
      - source: config/database.yml.example
        destination: config/database.yml
  - name: Deleting not needed files
    type: delete
    content:
      - path: Procfile
      - path: bower.json
  - name: Replacing variables with their respective content
    type: substitute
    content:
      - variable: gemset
        value: app_dev
      - variable: version
        value: ruby-2.2.4
  - name: Running bundle & using current gemset
    type: exec
    content:
      - command: rvm use .
      - command: gem install bundler
      - command: bundle
  - name: Setting up database, migrations and seeds
    type: exec
    content:
      - command: bundle exec rake db:create db:migrate
      - command: bundle exec rake db:seed
```

The config file should be placed at the root of the git repository to be detected and parsed by `cappuccino`. Once done, you can call the following command :
```
cappuccino -g git@github.com:username/reponame.git -b master
```

`-b master` is optional as the master branch is selected by default. Once the above command executed, `cappuccino` will clone the repository and apply the defined actions.

### Important
- If an action is a `substitution` and the variable name is `gemset`, cappuccino will search and find `[cappuccino-var-gemset]` in the repository and substitute it with related value.
- Both `substitution` and `replace` action types accept a `indent` key that informs `cappuccino` to indent the string or block by the desired number of spaces.
- The `path` key is optional but recommended; not defining it will force a [Depth-first Search Algorithm](https://en.wikipedia.org/wiki/Depth-first_search) on the whole repository.

### Display warnings
Cappuccino will find all files containing `[cappuccino-warning]` and display related file name and line number.
Adding `[cappuccino-warning]` to line 42 of `routes.rb` file will output upon config file parsing:
```
Please make sure to setup needed information located L-042 in config/routes.rb
```

### Currently available action types:
* _exec_, executes the given command with full list of arguments.
* _copy_, copies a file from source to destination.
* _move_, moves a file from source to destination.
* _replace_, replaces a string in a specified path or parses the whole repository.
* _substitute_, replaces cappuccino defined variables by their proper definition.
* _template_, takes a file from `.cappuccino` folder and copies it to repository.
* _delete_, deletes a file.

### Upcoming features:
* Automatic indexing of repository files, folders and cappuccino variables.
* Substitution with GPG encryption/decryption on the fly.
* Ability to call [Vault](https://www.hashicorp.com/vault.html) and retreive confidential information.
* Ability to call Amazon S3 for automatic bucket creation and credentials retrieval.
* Ability to call and execute a remote .cappuccino.yml config file.
* Compiled config file for semantic format analysis.
* Custom user defined types.

## License
```
The MIT License (MIT)

Copyright (c) 2016 CapSens S.A.S

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```
