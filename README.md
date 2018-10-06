## Description
  This is simple web application to analyze given url web page.  
  All is written in golang.

  Demonstration is below.

  ![result](https://github.com/16yuki0702/webanalyzer/blob/media/analyzer.gif)

## Usage
  To build or deploy this application, you must install docker first.  
  First of all clone this project.
``` bash
$ git clone this url
```

  And change your current directory to this project exists.
``` bash
$ cd /path/to/project
```

  The application can be build by below command.
``` bash
$ docker-compose up
```

  If you want to run this application not only localhost but also other environment,  
  you can modify environment file app.env.  
  application hostname and port depends on below environment value.
``` bash
ANALYZER_WEBSOCKET_HOST=localhost
ANALYZER_WEBSOCKET_PORT=8080
```
  defaule value is localhost:8080.
