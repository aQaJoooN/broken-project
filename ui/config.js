// API Configuration
const API_CONFIG = {
    baseUrl: 'http://localhost:8080',
    endpoints: {
        set: '/api/set',
        load: '/api/load',
        loadDb: '/api/load-db',
        metrics: '/metrics'
    }
};

// Helper function to get full URL
function getApiUrl(endpoint) {
    return API_CONFIG.baseUrl + API_CONFIG.endpoints[endpoint];
}
