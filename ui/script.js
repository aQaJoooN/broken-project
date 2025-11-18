document.getElementById('setForm').addEventListener('submit', async (e) => {
    e.preventDefault();
    
    const key = document.getElementById('key').value;
    const value = document.getElementById('value').value;
    const resultDiv = document.getElementById('result');
    
    try {
        const response = await fetch(getApiUrl('set'), {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ key, value })
        });
        
        const data = await response.json();
        
        if (data.success) {
            resultDiv.className = 'result success';
            resultDiv.textContent = `✓ ${data.message}`;
        } else {
            resultDiv.className = 'result error';
            resultDiv.textContent = `✗ ${data.message}`;
        }
        
        document.getElementById('setForm').reset();
        
    } catch (error) {
        resultDiv.className = 'result error';
        resultDiv.textContent = `✗ Error: ${error.message}`;
    }
});

document.getElementById('loadBtn').addEventListener('click', async () => {
    const loadBtn = document.getElementById('loadBtn');
    const loadResultDiv = document.getElementById('loadResult');
    
    loadBtn.disabled = true;
    loadBtn.textContent = 'Loading...';
    loadResultDiv.className = 'result info';
    loadResultDiv.textContent = '⏳ Starting Redis load test... This may take several minutes.';
    
    try {
        const response = await fetch(getApiUrl('load'), {
            method: 'GET'
        });
        
        const data = await response.json();
        
        if (data.success) {
            loadResultDiv.className = 'result success';
            loadResultDiv.textContent = `✓ ${data.message} - Check server logs for progress. This will take a few minutes.`;
            
            setTimeout(() => {
                loadBtn.disabled = false;
                loadBtn.textContent = 'Load on Redis';
            }, 5000);
        } else {
            loadResultDiv.className = 'result error';
            loadResultDiv.textContent = `✗ ${data.message}`;
            loadBtn.disabled = false;
            loadBtn.textContent = 'Load on Redis';
        }
        
    } catch (error) {
        loadResultDiv.className = 'result error';
        loadResultDiv.textContent = `✗ Error: ${error.message}`;
        loadBtn.disabled = false;
        loadBtn.textContent = 'Load on Redis';
    }
});

document.getElementById('loadDbBtn').addEventListener('click', async () => {
    const loadDbBtn = document.getElementById('loadDbBtn');
    const loadDbResultDiv = document.getElementById('loadDbResult');
    
    loadDbBtn.disabled = true;
    loadDbBtn.textContent = 'Loading...';
    loadDbResultDiv.className = 'result info';
    loadDbResultDiv.textContent = '⏳ Starting database load test... Opening 200 connections.';
    
    try {
        const response = await fetch(getApiUrl('loadDb'), {
            method: 'GET'
        });
        
        const data = await response.json();
        
        if (data.success) {
            loadDbResultDiv.className = 'result success';
            loadDbResultDiv.textContent = `✓ ${data.message} - Check server logs for progress.`;
            
            setTimeout(() => {
                loadDbBtn.disabled = false;
                loadDbBtn.textContent = 'Load on DataBase';
            }, 5000);
        } else {
            loadDbResultDiv.className = 'result error';
            loadDbResultDiv.textContent = `✗ ${data.message}`;
            loadDbBtn.disabled = false;
            loadDbBtn.textContent = 'Load on DataBase';
        }
        
    } catch (error) {
        loadDbResultDiv.className = 'result error';
        loadDbResultDiv.textContent = `✗ Error: ${error.message}`;
        loadDbBtn.disabled = false;
        loadDbBtn.textContent = 'Load on DataBase';
    }
});

// Initialize links on page load
document.addEventListener('DOMContentLoaded', () => {
    document.getElementById('link-set').href = getApiUrl('set');
    document.getElementById('link-set').textContent = getApiUrl('set');
    
    document.getElementById('link-load').href = getApiUrl('load');
    document.getElementById('link-load').textContent = getApiUrl('load');
    
    document.getElementById('link-load-db').href = getApiUrl('loadDb');
    document.getElementById('link-load-db').textContent = getApiUrl('loadDb');
    
    document.getElementById('link-metrics').href = getApiUrl('metrics');
    document.getElementById('link-metrics').textContent = getApiUrl('metrics');
});
