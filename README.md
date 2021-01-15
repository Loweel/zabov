# zabov

Tiny replacement for piHole DNS filter

Still Work in progress, usable.

Idea is to produce a very simple, no-web-interface , IP DNS blocker.

# INSTALL

Zabov requires golang 1.13 or later.

<pre>
git clone https://git.keinpfusch.net/Loweel/zabov.git
cd zabov
go get
go build -mod=vendor
</pre>

Then, edit config.json: please notice config.json must be in the same folder of the executable you run.


Just a few words about "singlefilters" and "doublefilters":

Data must be downloaded from URLs of blacklist mantainers.They may come in different formats.

There are two kinds of blacklists:

One is the format zabov calls "singlefilter", where we find a single column , full of domains:

<pre>
domain1.com
domain2.com
domain3.com
</pre>

The second is the format zabov calls "doublefilter" (a file in "/etc/hosts" format, to be precise), where there is an IP, usually localhost or 0.0.0.0 and then the domain:

<pre>
127.0.0.1 domain1.com
127.0.0.1 domain2.com
127.0.0.1 domain3.com
</pre>

This is why configuration file has two separated items.

Minimal config file should look like:

<pre>
{
    "zabov":{
        "port":"53", 
        "proto":"udp", 
        "ipaddr":"0.0.0.0",
        "cachettl": 1,
        "killfilettl": 12,
        "debug:"false"
    },
    "configs":{
        "default":{
            "upstream":"./dns-upstream.txt",
            "singlefilters":"./urls-domains.txt",
            "doublefilters":"./urls-hosts.txt", 
            "blackholeip":"127.0.0.1",
            "hostsfile":"./urls-local.txt"
        },
    }
}
</pre>

Global zabov settings:

- port is the port number. Usually is 53, you can change for docker, if you like
- proto is the protocol. Choices are "udp", "tcp", "tcp/udp"
- ipaddr is the port to listen to. Maybe empty, (which will result in listening to 0.0.0.0) to avoid issues with docker.
- cachettl: amount of time the cache is kept (in hours)
- killfilettl: refresh time for _killfiles_
- debug: if set to "true" Zabov prints verbose logs, such as config selection and single DNS requests

configs:
- contains multiple zabov configuration dictionaries. "default" configuration name is mandatory
- upstream: file containing all DNS we want to query :  each line in format IP:PORT
- singlefilters: name of the file  for blacklists following the "singlefilter" schema.(one URL per line)
- doublefilters: name of the file, for blacklists following the "doublefilter" schema.(one URL per line)
- blackholeip: IP address to return when the IP is banned. This is because you may want to avoid MX issues, mail loops on localhost, or you have a web server running on localhost
- hostsfile: path where you keep your local blacklistfile : this is in the format "singlefilter", meaning one domain per line, unlike hosts file.


Advanced configuration includes support for multiple configurations based on IP Source and timetables:
<pre>
{
    "zabov":{
        "port":"53", 
        "proto":"udp", 
        "ipaddr":"0.0.0.0",
        "cachettl": 1,
        "killfilettl": 12
    },
    "localresponder":{
        "responder":"192.168.178.1:53",
        "localdomain":"fritz.box"
    },
    "ipaliases":{
        "pc8":"192.168.178.29",
        "localhost":"127.0.0.1"
    },
    "ipgroups":[
        {
            "ips":["localhost", "::1", "192.168.178.30", "192.168.178.31", "pc8"],
            "cfg":"",
            "timetable":"tt_children"
        }
    ],
    "timetables":{
        "tt_children":{
            "tables":[{"times":"00:00-05:00;8:30-12:30;18:30-22:59", "days":"Mo;Tu;We;Th;Fr;Sa;Su"}],
            "cfgin":"children_restricted",
            "cfgout":"default"
        }
    },
    "configs":{
        "default":{
            "upstream":"./dns-upstream.txt",
            "singlefilters":"./urls-domains.txt",
            "doublefilters":"./urls-hosts.txt", 
            "blackholeip":"127.0.0.1",
            "hostsfile":"./urls-local.txt"
        },
        "children":{
            "upstream":"./dns-upstream-safe.txt",
            "singlefilters":"./urls-domains.txt",
            "doublefilters":"./urls-hosts.txt", 
            "blackholeip":"127.0.0.1",
            "hostsfile":"./urls-local.txt"
        },
        "children_restricted":{
            "upstream":"./dns-upstream-safe.txt",
            "singlefilters":"./urls-domains-restricted.txt",
            "doublefilters":"./urls-hosts-restricted.txt", 
            "blackholeip":"127.0.0.1",
            "hostsfile":"./urls-local.txt"
        }
    }
}
</pre>

localresponder:
  - allows to set a local DNS to respond for "local" domains. A domain name is handled as "local" if dosen't contains "." (dots) or if it ends with a well known prefix, such as ".local".
  Note: the cache is not used for local responder.
  - responder: is the local DNS server address in the IP:PORT format.
  - localdomain: is the suffix for local domain names. All domains ending with this prefix are resolved by local responder

ipaliases: a dictionary of IPs
  - each entry in this dictionary define a domain-alias name and his IP address. It works as replacement of  /etc/hosts file.
  - each entry is used by Zabov to resolve that names and to replace any value in the ipgroups.ips array.

timetables: a dictionary of timetable dictionaries
  - allow to define timetables in the format "time-ranges" and "days-of-week"
  - tables: contain an array of dictionaries, each defining a time rule.
    - each table is a dictinary containing "time" and "days" values
    - time: is a string in the form "start:time1-stop:time1;start:time2-stop:time2..."
    - days: is a string containing semicolon separated day names to apply the rule such as "Mo;Tu;We;Th;Fr"
      - days names are: "Mo", "Tu" "We", "Th", "Fr", "Sa", "Su"
      - empty value means all week-days
    You can define complex time rules using more than one entry in this dictionay
  - cfgin: is the name of the configuration to apply if current time is "inside" the timetable
  - cfgout: is the name of the configuration to apply if current time is "outside" the timetable
  
ipgroups: an array of ipgroup dictionaries
  - let you define a set of IP addresses that shall use a configuration other than "default"
  - ips: is an array of strings, each containing an ip address or a name defined in the "ipaliases" config branch
  - cfg: is a string containing the name of the configuration to be used for this group; ignored if timetable is also defined
  - timetable: is a string containing the name of the tiemtable to be aplied to this group


# DOCKER
Multistage Dockerfiles are provided for AMD64, ARMv7, ARM64V8

NOTE: you shall use TZ env var to change docker image timezone. TZ defaults to CET.

# TODO:

- ~~caching~~
- monitoring port


