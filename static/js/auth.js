document.addEventListener('DOMContentLoaded', function() {
    const registerForm = document.getElementById('register-form');
    if (registerForm) {
        registerForm.addEventListener('submit', async function(e) {
            e.preventDefault();
            
            const login = document.getElementById('login').value;
            const password = document.getElementById('password').value;
            const passwordConfirm = document.getElementById('password-confirm').value;

            if (login.length < 3) {
                showError("Логин должен содержать не менее 3 символов");
                return;
            }
            
            if (password.length < 5) {
                showError("Пароль должен содержать не менее 5 символов");
                return;
            }
            
            if (password !== passwordConfirm) {
                showError("Пароли не совпадают");
                return;
            }

            try {
                const response = await fetch('/api/v1/register', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify({ login, password })
                });
                
                const data = await response.json();
                
                if (!response.ok) {
                    if (response.status === 409) {
                        showError("Пользователь с таким логином уже существует");
                    } else {
                        showError(data.error || "Ошибка при регистрации");
                    }
                    return;
                }

                window.location.href = '/login?registered=true';
            } catch (error) {
                console.error("Ошибка при отправке запроса:", error);
                showError("Ошибка при отправке запроса на сервер");
            }
        });
    }

    const loginForm = document.getElementById('login-form');
    if (loginForm) {
        const urlParams = new URLSearchParams(window.location.search);
        if (urlParams.get('registered') === 'true') {
            showSuccess("Регистрация успешна! Теперь вы можете войти.");
        }
        
        loginForm.addEventListener('submit', async function(e) {
            e.preventDefault();
            
            const login = document.getElementById('login').value;
            const password = document.getElementById('password').value;

            try {
                const response = await fetch('/api/v1/login', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify({ login, password })
                });
                
                const data = await response.json();
                
                if (!response.ok) {
                    showError(data.error || "Ошибка при входе");
                    return;
                }

                if (data.token) {
                    localStorage.setItem('jwt_token', data.token);

                    document.cookie = `user_login=${login}; path=/`;
                }

                window.location.href = '/';
            } catch (error) {
                console.error("Ошибка при отправке запроса:", error);
                showError("Ошибка при отправке запроса на сервер");
            }
        });
    }
});

function showError(message) {
    const errorElement = document.getElementById('error-message');
    if (errorElement) {
        errorElement.textContent = message;
        errorElement.classList.remove('hidden');

        setTimeout(() => {
            errorElement.classList.add('hidden');
        }, 5000);
    }
}

function showSuccess(message) {
    const successElement = document.getElementById('success-message');
    if (successElement) {
        successElement.textContent = message;
        successElement.classList.remove('hidden');

        setTimeout(() => {
            successElement.classList.add('hidden');
        }, 5000);
    } else {
        const newSuccessElement = document.createElement('div');
        newSuccessElement.id = 'success-message';
        newSuccessElement.className = 'bg-green-500 text-white p-3 rounded-md mb-6';
        newSuccessElement.textContent = message;

        const container = document.querySelector('.login-container, .register-container');
        if (container) {
            container.insertBefore(newSuccessElement, container.firstChild);
            
            setTimeout(() => {
                newSuccessElement.style.display = 'none';
            }, 5000);
        }
    }
} 