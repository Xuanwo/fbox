
<a name="0.0.1"></a>
## 0.0.1 (2021-07-03)

### Bug Fixes

* Fix README
* Fix /files/ response when there are no files in the metdata db
* Fix a null pointer bug
* Fix height to 100%;
* Fix deps
* Fix download link
* Fix file icon location
* Fix paths for ui branch
* Fix bugs with put command
* Fix metadata endpoint
* Fix getRemoteMetadata uri used
* Fix bugs reconstructing shards
* Fix bugs with blob and upload handling
* Fix quick 'n dirty start cluster script

### Features

* Add some project infra borrowed from other projects
* Add a README and LICENSE
* Add a working Docker Swarm stack for the Mills DC
* Add Dockerfile to build Docker images
* Add extra fields for metdata objects to support directories, owner/group and file modes and permissions
* Add a sleep between starting master and nodes
* Add a Makefile with some build rules
* Add IP variable to start-cluster.sh to reduce typing
* Add multiple files in uploading
* Add list to ui
* Add put command
* Add sub-commands and cat command to directly download a file from nodes
* Add support for deleting blobs
* Add some missing Seek(s)
* Add 1s sleep between starting master and nodes
* Add single-node/local metadata persistence
* Add support for downloading files and reconstructing shards
* Add /files handler
* Add upload handler with reed solomon data shard and parity support with remote blob stores
* Add shells cript to start a cluster for testing
* Add fbox logo
* Add a prototype proxy service implementing join and nodes endpoints
* Add .gitkeep to data/

### Updates

* Update README
* Update deps

