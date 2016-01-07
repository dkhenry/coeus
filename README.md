# Coeus
A self hosted alternative to Atlas from Hashicorp

# Building Coeus
goxc is used to build coeus
```
goxc -bc='linux,darwin'
```

A suitable host enviroment can be created with a centos7 host and the
centos7.yml ansible script. A suitable hosts file will need to be
provided by the end user ( also you might need the yumrepo module ) 
```
ansible-playbook -i hosts centos7.yml
```

# Uploading new boxes.
Boxes can be uploaded via

```
curl --form "fileupload=@<BaseBox.box>"
http://localhost:8080/<namespace>/<name>/<version>/<provider>
```

That box will then be listed in the manifest and can be retrieved with

```
curl -o -
http://localhost:8080/<namespace>/boxes/<name>/<version>/<provider>.box
```

# Installing
If you are running something with selinux you will need to provide a
policy module to allow the system to work

```
sudo cat /var/log/audit/audit.log | grep nginx | grep denied |
audit2allow -M mynginx
sudo semodule -i mynginx.pp
```

This also requires you to have the Oracle VirtualBox Extensions
installed 
