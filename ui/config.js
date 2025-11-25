// API Configuration
const API_CONFIG = {
    baseUrl: 'http://localhost:8080',
    endpoints: {
        user: '/api/user',
        users: '/api/users',
        set: '/api/set',
        func1: '/api/func1',
        func2: '/api/func2',
        metrics: '/metrics'
    }
};

// Helper function to get full URL
function getApiUrl(endpoint) {
    return API_CONFIG.baseUrl + API_CONFIG.endpoints[endpoint];
}
