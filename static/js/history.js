async function loadExpressionHistory() {
    const userLogin = window.getCookieValue('user_login');
    
    if (!userLogin) {
        console.error("Ошибка: пользователь не авторизован");
        return;
    }
    
    try {
        const timestamp = new Date().getTime();
        
        const response = await fetch(`/api/v1/history?_t=${timestamp}`, {
            method: 'GET',
            headers: {
                'Content-Type': 'application/json',
                'X-User-Login': userLogin
            },
            credentials: 'include'
        });

        if (!response.ok) {
            const errorText = await response.text();
            console.error("Ошибка при загрузке истории:", response.status, errorText);
            return;
        }
        
        const data = await response.json();
        
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
            historyElement.innerHTML = '<div class="text-gray-500 text-center p-4">История пуста</div>';
            return;
        }


        
        const sortedExpressions = [...data.expressions];
        
        sortedExpressions.forEach(expr => {
            const expressionText = expr.text || expr.original || expr.expression || `Выражение #${expr.id}`;
            
            let resultValue = null;
            if (expr.result !== undefined && expr.result !== null) {
                resultValue = expr.result;
            } else if (expr.Result !== undefined && expr.Result !== null) {
                resultValue = expr.Result;
            }
            

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
    
    const historyElement = document.getElementById('history');
    if (historyElement) {
        historyElement.style.display = 'flex';
        historyElement.style.flexDirection = 'column-reverse';
        historyElement.style.overflowY = 'auto';
    }
    
    loadExpressionHistory();
}); 