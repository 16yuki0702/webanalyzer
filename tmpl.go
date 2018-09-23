package main

import (
	"html"
)

func indexPage() string {
	return `
<html>
	<head>
		<title>analyze web</title>
	</head>
	<body>
		<form action="" method=post>
			<input type=text name=value_post value="">
			<input type=submit name=submit value=submit>
		</form>
	</body>
</html>
`
}

func errorPage(err error) string {
	return `
<html>
    <head>
		<title>analyze web</title>
    </head>
    <body>
		<form action="" method=post>
			<input type=text name=value_post value="">
			<input type=submit name=submit value=submit>
		</form>
		<div>` + html.EscapeString(err.Error()) + `</div>
    </body>
</html>
`
}
