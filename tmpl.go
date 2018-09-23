package main

func indexPage() string {
	return `
<!DOCTYPE html>
<html>
	<head>
		<title>analyze web</title>
	</head>
	<script src="https://ajax.googleapis.com/ajax/libs/jquery/3.2.1/jquery.min.js"></script>
	<script type="text/javascript">
		var sock = null;
		var wsuri = "ws://localhost:8080/ws";

		window.onload = function() {
			sock = new WebSocket(wsuri);
			sock.onclose = function(e) {
				$('#results').append('<li style="color:red">connection closed ' + e.code + '</li>');
				$('#results').append('<li style="color:red">please restart server</li>');
			}
			sock.onmessage = function(e) {
				$('#results').append(e.data);
			}
		};
		function send() {
			url = document.getElementById('message').value;
			$("#results").empty();
			$('#results').append('<h2>analyze start : ' + url + '</h2>');
			sock.send(url);
		};
	</script>
	<body>
		<form onsubmit="return false;" method=post>
			<input id="message" type=text name="message" value="">
			<button onclick="send();">Send</button>
		</form>
		<ul id ="results"></ul>
	</body>
</html>
`
}