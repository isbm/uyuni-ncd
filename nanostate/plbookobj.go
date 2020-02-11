/*
Nanoplaybook implementation.
Playbook example:

----------------------------------------------
id: some-playbook
descr: This describes what this playbook is for.

state:
  maintain-database:
	- sshrunner:
	  - stop-db: systemctl stop postgresql.service
	  - backup-db: pg-backup /var/lib/pgsql/data /opt/backups/
	  - start-db: systemctl start postgresql.service
	- somemodule:
	  keyparam: valueparam
	  keyparam2: valueparam2
	  keyparam3: valueparam3

  some-other-group:
	- sshrunner:
	  - uptime: uptime
	  - id : cat /etc/machine-id
----------------------------------------------

In the example above, three fields are required:

- id
  This an ID of the playbook. It is used for the
  reporting at the end.

- descr:
  Description of the playbook. Also reporting.

- state:
  This is the entire tree of the playbook structure.
  It has twofold tree: a group IDs with a list of
  modules and the params or commands below.

Currently only one module is implemented: sshrunner,
which performs a series of synchronous commands in
the order they were placed for orchestration purposes.

In the nanoplaybook groups are asynchronous, but the
commands inside the groups are synchronous.
*/
package nanostate

type StateCommand struct {
	Id  string
	Cmd string
}

type StateGroup struct {
	Id    string
	Group []*StateCommand
}

type Playbook struct {
	groups []*StateGroup
}

func NewPlaybook() *Playbook {
	pb := new(Playbook)
	return pb
}

// Load playbook tree
func (pb *Playbook) Load(tree map[string]interface{}) *Playbook {
	pb.groups = nil
	pb.groups = make([]*StateGroup, 0)

	return pb
}
