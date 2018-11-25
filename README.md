# gssh  
<pre>
Usage: gssh [-t] [-p password] [-f file] [-v] [user@]hostname[:port] [command]
  -e    passing to pty
  -f string
        password file path
  -h    help
  -p string
        password
  -t    Force pseudo-tty allocation
  -v    Display Version
</pre>
## config
###  format
<pre>
[passwords]
user=password
user@host=password
user@host:port=password
</pre>

## Environment variable
### GSSH_PASSWORDFILE
gssh use this file,if specified.

### GSSH_PASSWORDS
    For example
      GSSH_PASSWORDS="user=password user@host=password"
</pre>

### ~/.gssh file
    gssh use ~/.gssh as config,if exist.
