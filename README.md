# Coeus
A self hosted alternative to Atlas from Hashicorp

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

