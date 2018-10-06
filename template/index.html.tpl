<!DOCTYPE html>
<html>
	<head>
		<title>Web Analyzer</title>
	</head>
	<link rel="stylesheet" href="//maxcdn.bootstrapcdn.com/bootstrap/3.2.0/css/bootstrap.min.css">
	<script src="//ajax.googleapis.com/ajax/libs/jquery/3.2.1/jquery.min.js"></script>
	<script src="//maxcdn.bootstrapcdn.com/bootstrap/3.2.0/js/bootstrap.min.js"></script>
	<script type="text/javascript">
		var sock = null;
		var wsuri = "ws://{{.WebSocketHost}}:{{.WebSocketPort}}/ws";
		const SUCCESS = 0;
		const FAILURE = 1;
		const COMPLETE = 2;

		$(function(){
			sock = new WebSocket(wsuri);
			sock.onclose = function(e) {
				$('#results').append('<li class="list-group-item list-group-item-danger">Server Closed</li>');
				$('#results').append('<li class="list-group-item list-group-item-danger">Error Code : ' + e.code + '</li>');
				$('#results').append('<li class="list-group-item list-group-item-danger">Please Restart Server</li>');
			}
			sock.onmessage = function(e) {
				response = JSON.parse(e.data)
				if (response.Status == SUCCESS) {
					$('#results').append('<li class="list-group-item list-group-item-success">' + response.Result + '</li>');
				} else if (response.Status == FAILURE) {
					$('#results').append('<li class="list-group-item list-group-item-danger">' + response.Result + '</li>');
				} else if (response.Status == COMPLETE) {
					$('#results').append('<li class="list-group-item list-group-item-info">' + response.Result + '</li>');
				}
			}
		});
		function send() {
			url = $('#message').val();
			$('#results').empty();
			$('#results').append('<li class="list-group-item list-group-item-info">analyze start : ' + url + '</li>');
			sock.send(url);
		};
	</script>
	<body>
		<div class="container">
			<h3 class="h3">Web Analyzer</h3>
			<form onsubmit='return false;' method=post>
				<div class="form-group">
					<input id='message' class="form-control">
				</div>
				<button class="btn btn-default" onclick='send();'>Send</button>
			</form>
			<ul id ='results' class="list-group"></ul>
		</div>
	</body>
</html>
