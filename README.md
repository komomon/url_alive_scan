
## url_alive_scan

一个go编写的多协程并发批量url存活检测工具，并发数默认根据cpu决定。

由于httpx探活准确性并不是特别高，存在漏报的情况，故写了这个工具。

实测和httpx各有千秋，结果总量会比httpx多一些，重点还是协程数不要开太高，5-20为宜。


### Usage
```cmd
Usage of ./url_alive_scan:
-o string
    the result file (default "time-inputfilename.txt")
-i string
    the file of the targets (default "urls.txt")
-retry int
    after failure, the number of request attempts (default 1)
-t int
    the threads of the program (default 20)
```


### example

```cmd
url_alive_scan.exe -i .\urls2.txt
                Url_alive_scan 1.0 #dev
                                        by komomon
[+]https://[2:c080:1xxxxxx6:59b0]/test.html 404 Not Found 404 546
[+]https://bdhxxxx.com/test.html 404 Not Found 404 562
[+]https://axxxx.com/test.html 404 Not Found 404 546
[+]https://bidxxxxxv.com/test.html 404 Not Found 404 546
[+]https://aaq.com/1.html 域名未配置 530 4152
[+]https://aaaa.com/test.html 404 Not Found - 综合网 404 481
[+] Outputfile: 2024-01-09-16-38-29-urls2.txt
[+] Alive url number: 6

```









