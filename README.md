# gssh  
*Install*  
$ go get -u github.com/golang/dep/cmd/dep  
$ cd $GOPATH/src  
$ git clone https://github.com/takemxn/gssh  
$ cd gssh  
$ dep ensure  
$ go build  
$ go install  

*How to use*  
Usage : gssh [-t] [-p password] [-f file] [user@]hostname[:port] [command]  

ex:  
$ gssh -p hogepassword hoge@hoge.com  
