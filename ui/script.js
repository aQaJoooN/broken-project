// Wait for DOM to be ready
function initializeApp() {
    // User form handler
    document.getElementById('userForm').addEventListener('submit', async (e) => {
        e.preventDefault();

        const firstName = document.getElementById('firstName').value;
        const lastName = document.getElementById('lastName').value;
        const age = parseInt(document.getElementById('age').value);
        const maritalStatus = document.getElementById('maritalStatus').value === 'true';
        const resultDiv = document.getElementById('userResult');

        try {
            const response = await fetch(getApiUrl('user'), {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ 
                    first_name: firstName,
                    last_name: lastName,
                    age: age,
                    marital_status: maritalStatus
                })
            });

            const data = await response.json();

            if (data.success) {
                resultDiv.className = 'result success';
                resultDiv.textContent = `✓ ${data.message} (User ID: ${data.user_id})`;
            } else {
                resultDiv.className = 'result error';
                resultDiv.textContent = `✗ ${data.message}`;
            }

            document.getElementById('userForm').reset();

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

    // Get Users button handler
    document.getElementById('getUsersBtn').addEventListener('click', async () => {
        const getUsersBtn = document.getElementById('getUsersBtn');
        const getUsersResultDiv = document.getElementById('getUsersResult');
        const usersListDiv = document.getElementById('usersList');

        getUsersBtn.disabled = true;
        getUsersBtn.textContent = 'Loading...';
        getUsersResultDiv.className = 'result info';
        getUsersResultDiv.textContent = '⏳ Fetching users...';
        usersListDiv.innerHTML = '';

        try {
            const response = await fetch(getApiUrl('users'), {
                method: 'GET'
            });

            const data = await response.json();

            if (data.success) {
                getUsersResultDiv.className = 'result success';
                getUsersResultDiv.textContent = `✓ ${data.message}`;

                if (data.users && data.users.length > 0) {
                    let usersHTML = '<div class="users-table"><table><thead><tr><th>User ID</th><th>Data</th></tr></thead><tbody>';
                    data.users.forEach(user => {
                        usersHTML += `<tr><td>${user.user_id || 'N/A'}</td><td>${user.data || JSON.stringify(user)}</td></tr>`;
                    });
                    usersHTML += '</tbody></table></div>';
                    usersListDiv.innerHTML = usersHTML;
                } else {
                    usersListDiv.innerHTML = '<p class="no-users">No users found.</p>';
                }

                getUsersBtn.disabled = false;
                getUsersBtn.textContent = 'Get Users';
            } else {
                getUsersResultDiv.className = 'result error';
                getUsersResultDiv.textContent = `✗ ${data.message}`;
                getUsersBtn.disabled = false;
                getUsersBtn.textContent = 'Get Users';
            }

        } catch (error) {
            getUsersResultDiv.className = 'result error';
            getUsersResultDiv.textContent = `✗ Error: ${error.message}`;
            getUsersBtn.disabled = false;
            getUsersBtn.textContent = 'Get Users';
        }
    });

    // Initialize links
    document.getElementById('link-user').href = getApiUrl('user');
    document.getElementById('link-user').textContent = getApiUrl('user');
    
    document.getElementById('link-users').href = getApiUrl('users');
    document.getElementById('link-users').textContent = getApiUrl('users');
    
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
