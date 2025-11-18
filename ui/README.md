# UI Configuration

## Changing API URL

To change the API base URL, edit the `config.js` file:

```javascript
const API_CONFIG = {
    baseUrl: 'http://localhost:8080',  // Change this to your API URL
    endpoints: {
        set: '/api/set',
        load: '/api/load',
        loadDb: '/api/load-db',
        metrics: '/metrics'
    }
};
```

### Examples:

**For production:**
```javascript
baseUrl: 'https://api.example.com'
```

**For different port:**
```javascript
baseUrl: 'http://localhost:9000'
```

**For different host:**
```javascript
baseUrl: 'http://192.168.1.100:8080'
```

## Files

- `config.js` - API configuration (edit this to change URLs)
- `index.html` - Main HTML page
- `script.js` - JavaScript functionality
- `styles.css` - Styling

## Deployment

After changing `config.js`, rebuild the Docker container:

```bash
docker-compose up --build ui
```

Or if serving statically, just update the file and refresh the browser.
