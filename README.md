## Cappuccino
Cappuccino is a [Go](https://golang.org) project that aims to help developers avoid doing repetitive tasks by defining a structured `.cappuccino.yml` config file.

Because robots should do the hard work, a typical config file can contain from few to dozens of actions requiring executing commands, renaming, searching and replacing, generating from templates and more. Cappuccino is written in Go and is thus executable on all plateforms, without a need for an `LLVM` or an interpreter.

### Installation
```
sudo curl -sSo /usr/bin/cappuccino https://raw.githubusercontent.com/CapSens/cappuccino/master/cappuccino && sudo chmod 777 /usr/bin/cappuccino
```

### Example of config file
```yaml
engine: cappuccino
actions:
  - name: Copying database config file
    type: copy
    content:
      - source: config/database.yml.example
        destination: config/database.yml
  - name: Moving Licence file to Licence.back
    type: move
    content:
      - source: LICENSE
        destination: LICENSE.back
  - name: Deleting not needed files
    type: delete
    content:
      - path: Procfile
      - path: bower.json
  - name: Setting up database, migrations and seeds
    type: exec
    content:
      - command: bundle exec rake db:create db:migrate
      - command: bundle exec rake db:seed
```

### Currently available action types:
* _exec_, executes the given command with full list of arguments.
* _copy_, copies a file from source to destination.
* _move_, moves a file from source to destination.
* _replace_, replaces a string in a specified path or by parsing the whole repository.
* _substitute_, replaces cappuccino defined variables by their proper definition.
* _delete_, deletes a file.

### Upcoming features:
* Substitution with GPG encryption/decryption on the fly.
* Ability to call [Vault](https://www.hashicorp.com/vault.html) and retreive confidential information.
* Compilable config file for semantic format analysis.
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
