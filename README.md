# Installing and Upgrading

    $ GOBIN=~/.local/bin go get -u github.com/simon3z/rhusers

# Usage

Parameters:

    $ rhusers -h
    Usage of rhusers:
      -b string
            base dn for search queries (default "ou=users,dc=redhat,dc=com")
      -g    google sheets format
      -q string
            ldap query string (default "(uid={})")
      -s string
            ldap server address and port (default "ldap.corp.redhat.com:389")

Example:

    $ echo fsimonce | rhusers -g
