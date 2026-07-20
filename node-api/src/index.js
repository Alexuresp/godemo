const express = require('express');

const app = express();

const config = {
  appName: process.env.APP_NAME || 'node-api',
  port: parseInt(process.env.PORT || '3000', 10),
  greeting: process.env.GREETING || 'Hello',
  secret: process.env.API_SECRET || '',
};

app.get('/healthz', (req, res) => {
  res.json({ status: 'ok', app: config.appName });
});

app.get('/', (req, res) => {
  res.json({
    message: `${config.greeting} from ${config.appName}`,
    secret: config.secret ? '***' : '(empty)',
  });
});

app.listen(config.port, () => {
  console.log(`listening on ${config.port}`);
});
