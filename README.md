# Installing and Upgrading

    $ GOBIN=~/.local/bin GO111MODULE=on go get github.com/simon3z/rhusers

# Usage

Parameters:

    $ rhusers -h
    Usage of rhusers:
      -b string
            base dn for search queries (default "ou=users,dc=redhat,dc=com")
      -g    google sheets format
      -j string
            jira user name
      -q string
            ldap query string (default "(uid={})")
      -s string
            ldap server address and port (default "ldap.corp.redhat.com:389")
      -z string
            jira server url (default "https://issues.redhat.com")

Example:

    $ echo fsimonce | rhusers -g
