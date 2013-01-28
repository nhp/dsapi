#!/bin/bash
ICONVBIN='/usr/bin/iconv' # path to iconv binary
dt=$(date +%s)
for f in /var/www/vhosts/dsapi/htdocs/data/*.xml
do
    if test -f $f
    then
        echo -e "\nConverting $f"
        /bin/mv $f $f.latin.$dt
        $ICONVBIN -f latin1 -t utf-8 $f.latin.$dt > $f
    else
        echo -e "\nSkipping $f - not a regular file";
    fi
done
/usr/local/go/bin/go run /var/www/vhosts/dsapi/import.go
/bin/rm /var/www/vhosts/dsapi/htdocs/data/*.xml
