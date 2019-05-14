
Usage
=====

* `gaffer -help` 
    Return command-line options

* `gaffer -version`
    Return information about the build of the tool

* `gaffer` 
    Return list of services and instances

* `gaffer /` 
    Return list of executables

* `gaffer @` 
    Return list of groups

* `gaffer _` 
    Return list of service types

* `gaffer <service>|@<group>|<instance>|_<dns-sd>`
    Return information on a service, group, instance or DNS-SD service records

* `gaffer /<exec> add`
    Add a executable

* `gaffer /<exec> add name=<service> mode=(auto|manual) run=<duration> idle=<duration> groups=@<groups> instance_count=<uint>`
    Add a executable, setting options

* `gaffer <service>|@<group> rm`
    Remove a service or group

* `gaffer <service>|@<group> flags (<key>|<key>=<value>)...`
    Set flags for a service or group. For keys without values, these are
    assumed to be boolean true.

* `gaffer @<group> env (<key>|<key>=<value>)...`
    Set environment parameters for a service or group. For keys without
    values, these are assumed to be empty strings.

* `gaffer <service>|@<group> start`
    Start instances for a service or group. Will tail the instance(s) which are started,
    press CTRL+C to stop. Use "-notail" option to return immediately.

* `gaffer <service>|@<group>|<instance> stop`
    Stop instance, instances for a service or group. Will tail the instance(s) which are started,
    press CTRL+C to stop. Use "-notail" option to return immediately.

* `gaffer <service> set name=<service> mode=(auto|manual) run=<duration> idle=<duration> groups=@<groups> instance_count=<uint>`
    Edit a service

* `gaffer @<group> set name=<service>`
    Edit a group name

* `gaffer <service> (disable|enable)`
    Set instance count to 0 or 1

* `gaffer <service>|@<group>|<instance> tail`
    Tail the log for an instance, service or group (press CTRL+C to end)

<group> starts with an amperstand character, for example "@rpc"
<grouplist> starts with an amperstand character, and comma-separated list, ie "@rpc,ssl,debug"
<instance> is a non-zero positive number, for example "4567"
<service> is an identifier starting with a letter, for example "remotes-service"
<dns-sd> records start with an underscore, for example "_gopi._tcp"
<exec> start with the path separator character, usually '/'

Here are some special group names:

@rpc - Adds -rpc.port=<port> -rpc.sslkey=<key> -rpc.sslcert=<cert> onto the command line,
  where they are provided to the gaffer service. The <port> is automatically generated to be
  an unused port.
@debug - Adds -debug
@debug2 - Adds -debug -verbose
@info - Adds -verbose
