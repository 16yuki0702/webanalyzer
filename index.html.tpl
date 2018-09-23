<!DOCTYPE html>
<html>
	<head>
		<title>analyze web</title>
	</head>
	<script src="https://ajax.googleapis.com/ajax/libs/jquery/3.2.1/jquery.min.js"></script>
	<script type="text/javascript">
		var sock = null;
		var wsuri = "ws://{{.WebSocketHost}}/ws";
		const SUCCESS = 0;
		const FAILURE = 1;
		const COMPLETE = 2;

		window.onload = function() {
			sock = new WebSocket(wsuri);
			sock.onclose = function(e) {
				$('#results').append('<li style="color:red">connection closed ' + e.code + '</li>');
				$('#results').append('<li style="color:red">please restart server</li>');
			}
			sock.onmessage = function(e) {
				response = JSON.parse(e.data)
				if (response.Status == SUCCESS) {
					$('#results').append('<li>' + response.Result + '</li>');
				} else if (response.Status == FAILURE) {
					$('#results').append('<li style="color:red">' + response.Result + '</li>');
				} else if (response.Status == COMPLETE) {
					$('#results').append('<h2>' + response.Result + '</h2>');
				}
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
			<input id="message" type=text name="message" value="" size=64>
			<button onclick="send();">Send</button>
		</form>
		<ul id ="results"></ul>
	</body>
</html>
