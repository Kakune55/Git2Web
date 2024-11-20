
## docker部署
~~~ bash
docker run -d -p 80:8080 -p 8081:8081 -v ./git2web/conf:/root/conf -v ./git2web/logs:/root/logs --restart always --name git2web kakune55/git2web
~~~