async function loadExpressionHistory() {
    const userLogin = window.getCookieValue('user_login');
    console.log("Загрузка истории для пользователя:", userLogin);
    
    if (!userLogin) {
        console.error("Ошибка: пользователь не авторизован");
        return;
    }
    
    try {
        console.log(`Отправка запроса на /api/v1/history с заголовком X-User-Login: ${userLogin}`);
        
        const timestamp = new Date().getTime();
        
        const response = await fetch(`/api/v1/history?_t=${timestamp}`, {
            method: 'GET',
            headers: {
                'Content-Type': 'application/json',
                'X-User-Login': userLogin
            },
            credentials: 'include'
        });
        
        console.log(`Получен ответ со статусом: ${response.status}`);
        
        if (!response.ok) {
            const errorText = await response.text();
            console.error("Ошибка при загрузке истории:", response.status, errorText);
            return;
        }
        
        const data = await response.json();
        console.log("Полученные данные истории:", data);
        
        const historyElement = document.getElementById('history');
        if (!historyElement) {
            console.error("Элемент 'history' не найден в DOM");
            return;
        }
        
        historyElement.style.display = 'flex';
        historyElement.style.flexDirection = 'column-reverse';
        historyElement.style.overflowY = 'auto';
        
        historyElement.innerHTML = '';
        
        if (!data.expressions || data.expressions.length === 0) {
            console.log("История выражений пуста");
            historyElement.innerHTML = '<div class="text-gray-500 text-center p-4">История пуста</div>';
            return;
        }
        
        console.log("Найдено выражений:", data.expressions.length);
        
        const sampleExpr = data.expressions[0];
        console.log("Пример структуры данных выражения:", sampleExpr);
        
        const sortedExpressions = [...data.expressions];
        
        sortedExpressions.forEach(expr => {
            const expressionText = expr.text || expr.original || expr.expression || `Выражение #${expr.id}`;
            
            let resultValue = null;
            if (expr.result !== undefined && expr.result !== null) {
                resultValue = expr.result;
            } else if (expr.Result !== undefined && expr.Result !== null) {
                resultValue = expr.Result;
            }
            
            console.log(`Добавляю в историю: ${expressionText} с результатом: ${resultValue}, статус: ${expr.status}`);
            
            if (expr.status === "COMPLETED" && resultValue !== null) {
                addToHistory(expressionText, resultValue);
            } 
            else if (resultValue !== null) {
                addToHistory(expressionText, resultValue);
            }
        });
    } catch (error) {
        console.error('Ошибка при загрузке истории:', error);
    }
}

function addToHistory(expression, result) {
    const historyElement = document.getElementById('history');
    if (!historyElement) return;
    
    console.log(`Добавление в историю: ${expression} = ${result}`);
    
    const historyItem = document.createElement('div');
    historyItem.className = 'history-item opacity-0 w-full mx-auto text-center';
    historyItem.innerHTML = `
        <div class="text-sm text-white/80 mb-1">${expression}</div>
        <div class="text-xl font-bold text-white">${result}</div>
    `;
    
    historyElement.append(historyItem);
    
    if (window.gsap) {
        gsap.to(historyItem, {
            opacity: 1,
            y: 20,
            duration: 0.5
        });
    } else {
        historyItem.style.opacity = 1;
    }
    
    historyElement.scrollTop = 0;
}

function useExpression(expression) {
    const input = document.getElementById('expression');
    if (input) {
        input.value = expression;
        if (typeof evaluateExpression === 'function') {
            evaluateExpression();
        }
    }
}

document.addEventListener('DOMContentLoaded', function() {
    console.log("Загрузка истории выражений при загрузке страницы");
    
    const historyElement = document.getElementById('history');
    if (historyElement) {
        historyElement.style.display = 'flex';
        historyElement.style.flexDirection = 'column-reverse';
        historyElement.style.overflowY = 'auto';
    }
    
    loadExpressionHistory();
}); 