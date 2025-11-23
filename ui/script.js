// Wait for DOM to be ready
function initializeApp() {
    // Set form handler
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

    // Func1 button handler
    document.getElementById('func1Btn').addEventListener('click', async () => {
        const func1Btn = document.getElementById('func1Btn');
        const func1ResultDiv = document.getElementById('func1Result');

        func1Btn.disabled = true;
        func1Btn.textContent = 'Loading...';
        func1ResultDiv.className = 'result info';
        func1ResultDiv.textContent = '⏳ Starting Func 1... This may take several minutes. Check server logs for progress.';

        try {
            const response = await fetch(getApiUrl('func1'), {
                method: 'GET'
            });

            const data = await response.json();

            if (data.success) {
                func1ResultDiv.className = 'result success';
                func1ResultDiv.textContent = `✓ ${data.message} - This will take a few minutes.`;

                setTimeout(() => {
                    func1Btn.disabled = false;
                    func1Btn.textContent = 'Func 1';
                }, 5000);
            } else {
                func1ResultDiv.className = 'result error';
                func1ResultDiv.textContent = `✗ ${data.message}`;
                func1Btn.disabled = false;
                func1Btn.textContent = 'Func 1';
            }

        } catch (error) {
            func1ResultDiv.className = 'result error';
            func1ResultDiv.textContent = `✗ Error: ${error.message}`;
            func1Btn.disabled = false;
            func1Btn.textContent = 'Func 1';
        }
    });

    // Func 2 button handler
    document.getElementById('func2Btn').addEventListener('click', async () => {
        const func2Btn = document.getElementById('func2Btn');
        const func2ResultDiv = document.getElementById('func2Result');

        func2Btn.disabled = true;
        func2Btn.textContent = 'Loading...';
        func2ResultDiv.className = 'result info';
        func2ResultDiv.textContent = '⏳ Starting Func 2... This may take several minutes. Check server logs for progress.';

        try {
            const response = await fetch(getApiUrl('func2'), {
                method: 'GET'
            });

            const data = await response.json();

            if (data.success) {
                func2ResultDiv.className = 'result success';
                func2ResultDiv.textContent = `✓ ${data.message} - This will take a few minutes.`;

                setTimeout(() => {
                    func2Btn.disabled = false;
                    func2Btn.textContent = 'Func 2';
                }, 5000);
            } else {
                func2ResultDiv.className = 'result error';
                func2ResultDiv.textContent = `✗ ${data.message}`;
                func2Btn.disabled = false;
                func2Btn.textContent = 'Func 2';
            }

        } catch (error) {
            func2ResultDiv.className = 'result error';
            func2ResultDiv.textContent = `✗ Error: ${error.message}`;
            func2Btn.disabled = false;
            func2Btn.textContent = 'Func 2';
        }
    });

    // Initialize links
    document.getElementById('link-set').href = getApiUrl('set');
    document.getElementById('link-set').textContent = getApiUrl('set');

    document.getElementById('link-func1').href = getApiUrl('func1');
    document.getElementById('link-func1').textContent = getApiUrl('func1');

    document.getElementById('link-func2').href = getApiUrl('func2');
    document.getElementById('link-func2').textContent = getApiUrl('func2');

    document.getElementById('link-metrics').href = getApiUrl('metrics');
    document.getElementById('link-metrics').textContent = getApiUrl('metrics');
}

// Run when DOM is ready
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', initializeApp);
} else {
    initializeApp();
}
