package auth

const noAuthContent = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Authorization Successful</title>
    <script type="text/javascript">
        function closeWindow() {
            window.close();
        }
    </script>
</head>
<body>
    <h1>Not authenticated</h1>
    <button onclick="closeWindow()">Start</button>
</body>
</html>
`

const authedContent = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Authorization Successful</title>
    <script type="text/javascript">
        function closeWindow() {
            window.close();
        }
    </script>
</head>
<body>
    <h1>Authorization Successful</h1>
    <button onclick="closeWindow()">Close</button>
</body>
</html>
`
