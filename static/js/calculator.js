(function() {
    window.getCookieValue = function(name) {
        const value = `; ${document.cookie}`;
        const parts = value.split(`; ${name}=`);
        if (parts.length === 2) return parts.pop().split(';').shift();
        return null;
    };
})();

class Calculator3D {
    constructor() {
        this.scene = new THREE.Scene();
        this.renderer = new THREE.WebGLRenderer({
            antialias: true,
            alpha: false
        });
        document.body.style.background = '#000000';
        this.camera = new THREE.PerspectiveCamera(75, window.innerWidth / window.innerHeight, 0.1, 1000);
        this.floatingNumbers = [];
        this.stars = [];
        this.backgroundPixels = [];
        this.loadingAnimation = null;
        this.infinityImage = null;
        this.infinityTime = 0;
        this.targetPosition = new THREE.Vector3(0, -2, 0);
        this.maxStars = 20;
        this.spawnPoints = [
            new THREE.Vector3(-20, 20, 0),
            new THREE.Vector3(20, 20, 0),
            new THREE.Vector3(0, 25, 0),
        ];
        
        const fontLoader = new THREE.FontLoader();
        fontLoader.load('https://threejs.org/examples/fonts/helvetiker_regular.typeface.json', 
            font => this.font = font
        );
        
        this.init();
        this.backgroundPixels = this.createBackgroundPattern();
        this.startStarSpawning();
        this.createInfinityImage();
        this.animate();
    }

    init() {
        this.renderer.setSize(window.innerWidth, window.innerHeight);
        document.getElementById('canvas-container').appendChild(this.renderer.domElement);
        
        this.camera.position.z = 15;

        const light = new THREE.DirectionalLight(0xffffff, 1);
        light.position.set(0, 0, 1);
        this.scene.add(light);
        
        const ambientLight = new THREE.AmbientLight(0x404040);
        this.scene.add(ambientLight);
    }

    createStarGeometry() {
        const points = [];
        for (let i = 0; i < 5; i++) {
            const angle = (i * Math.PI * 2) / 5;
            points.push(new THREE.Vector2(Math.cos(angle) * 1.0, Math.sin(angle) * 1.0));
            const innerAngle = angle + Math.PI / 5;
            points.push(new THREE.Vector2(Math.cos(innerAngle) * 0.3, Math.sin(innerAngle) * 0.3));
        }
        const shape = new THREE.Shape(points);
        return new THREE.ExtrudeGeometry(shape, {
            depth: 0.2,
            bevelEnabled: false
        });
    }

    startStarSpawning() {
        for (let i = 0; i < 3; i++) {
            this.createStar();
        }

        const spawnStar = () => {
            if (this.stars.length < this.maxStars) {
                this.createStar();
            }
            setTimeout(spawnStar, 2000 + Math.random() * 1000);
        };
        spawnStar();
    }

    createStar() {
        const starGeometry = this.createStarGeometry();
        const spawnPoint = this.spawnPoints[Math.floor(Math.random() * this.spawnPoints.length)];
        const randomOffset = Math.random() * 2 - 1;
        const initialColor = new THREE.Color(this.getRandomRainbowColor());
        const material = new THREE.MeshPhongMaterial({
            color: initialColor,
            emissive: initialColor,
            emissiveIntensity: 0.5,
            transparent: true,
            opacity: 0.8,
            side: THREE.DoubleSide
        });
        
        const star = new THREE.Mesh(starGeometry, material);
        star.position.copy(spawnPoint);
        star.position.x += randomOffset * 8;
        star.position.y += Math.abs(randomOffset) * 5;
        
        this.scene.add(star);
        this.stars.push({
            mesh: star,
            velocity: new THREE.Vector3(
                (this.targetPosition.x - star.position.x) * 0.008,
                (this.targetPosition.y - star.position.y) * 0.008,
                0
            ),
            rotationSpeed: {
                x: (Math.random() - 0.5) * 0.02,
                y: (Math.random() - 0.5) * 0.02,
                z: (Math.random() - 0.5) * 0.02
            },
            color: {
                current: initialColor,
                target: new THREE.Color(this.getRandomRainbowColor()),
                step: 0.02
            },
            hitCalculator: false
        });
    }

    getRandomRainbowColor() {
        const rainbow = [
            0x2222FF,
            0x22FFFF,
            0xFF22FF,
            0xFF2222,
            0x22FF22,
            0xFFFF22
        ];
        return rainbow[Math.floor(Math.random() * rainbow.length)];
    }

    createStarExplosion(position, color) {
        const smallStarGeometry = this.createStarGeometry();
        
        for (let i = 0; i < 8; i++) {
            const starColor = new THREE.Color(color);
            const material = new THREE.MeshPhongMaterial({
                color: starColor,
                emissive: starColor,
                emissiveIntensity: 0.5,
                transparent: true,
                opacity: 0.8,
                side: THREE.DoubleSide
            });
            
            const smallStar = new THREE.Mesh(smallStarGeometry, material);
            smallStar.position.copy(position);
            smallStar.scale.setScalar(0.3);
            
            const angle = (i * Math.PI * 2) / 8;
            const speed = 0.3;
            const velocity = {
                x: Math.cos(angle) * speed * 0.5,
                y: Math.sin(angle) * speed * 0.5,
                z: Math.random() * 0.1 - 0.05
            };
            
            this.scene.add(smallStar);
            this.stars.push({
                mesh: smallStar,
                velocity,
                rotationSpeed: {
                    x: Math.random() * 0.1,
                    y: Math.random() * 0.1,
                    z: Math.random() * 0.1
                },
                lifespan: 100,
                isExplosion: true,
                color: {
                    current: starColor,
                    target: new THREE.Color(this.getRandomRainbowColor()),
                    step: 0.05
                }
            });
        }
    }

    addFloatingNumber(number) {
        if (!this.font) return;
        const geometry = new THREE.TextGeometry(number.toString(), {
            font: this.font,
            size: 3,
            height: 0.5,
        });
        
        geometry.computeBoundingBox();
        const centerOffset = new THREE.Vector3(
            -(geometry.boundingBox.max.x - geometry.boundingBox.min.x) / 2,
            -(geometry.boundingBox.max.y - geometry.boundingBox.min.y) / 2,
            -(geometry.boundingBox.max.z - geometry.boundingBox.min.z) / 2
        );
        geometry.translate(centerOffset.x, centerOffset.y, centerOffset.z);
        
        const material = new THREE.MeshPhongMaterial({
            color: this.getRandomRainbowColor(),
            emissive: this.getRandomRainbowColor(),
            transparent: true,
            opacity: 0.8,
        });
        
        const mesh = new THREE.Mesh(geometry, material);
        const spawnPositions = [
            { x: -15, y: 25 },
            { x: 0, y: 30 },
            { x: 15, y: 25 }
        ];
        const pos = spawnPositions[Math.floor(Math.random() * spawnPositions.length)];
        mesh.position.x = pos.x + (Math.random() - 0.5) * 5;
        mesh.position.y = pos.y;
        mesh.position.z = (Math.random() - 0.5) * 5;
        mesh.rotation.x = 0;
        mesh.rotation.y = Math.PI;
        
        this.scene.add(mesh);
        this.floatingNumbers.push({
            mesh,
            velocity: new THREE.Vector3(
                (Math.random() - 0.5) * 0.05,
                -0.08,
                0
            ),
            rotationSpeed: 0.01
        });
    }

    addFloatingOperator(operator) {
        if (!this.font) return;
        const geometry = new THREE.TextGeometry(operator, {
            font: this.font,
            size: 4,
            height: 0.8,
        });
        
        const material = new THREE.MeshPhongMaterial({
            color: 0xf43f5e,
            transparent: true,
            opacity: 0.8,
            emissive: 0xf43f5e,
            emissiveIntensity: 0.5
        });
        
        const mesh = new THREE.Mesh(geometry, material);
        mesh.position.x = (Math.random() - 0.5) * 20;
        mesh.position.y = 15;
        mesh.position.z = (Math.random() - 0.5) * 10;
        
        this.scene.add(mesh);
        this.floatingNumbers.push({
            mesh,
            velocity: new THREE.Vector3(0, -0.05, 0),
            rotationSpeed: 0.01
        });
    }

    startLoadingAnimation() {
        const geometry = new THREE.TorusGeometry(2, 0.5, 16, 100);
        const material = new THREE.MeshPhongMaterial({
            color: 0x6366f1,
            transparent: true,
            opacity: 0.8,
        });
        this.loadingAnimation = new THREE.Mesh(geometry, material);
        this.loadingAnimation.position.set(0, 0, -5);
        this.scene.add(this.loadingAnimation);
    }

    animate() {
        requestAnimationFrame(() => this.animate());
        
        this.colorAnimationTime += this.colorChangeSpeed;
        
        this.backgroundPixels.forEach(pixel => {
            pixel.mesh.position.x += pixel.speed * pixel.direction;
            
            const maxOffset = Math.abs(pixel.initialX) + 40;
            if (Math.abs(pixel.mesh.position.x) > maxOffset) {
                pixel.mesh.position.x = pixel.initialX;
            }
            
            const colorIndex = Math.floor(this.colorAnimationTime) % this.colorSteps.length;
            const nextColorIndex = (colorIndex + 1) % this.colorSteps.length;
            const fraction = this.colorAnimationTime % 1;
            
            const currentColor = new THREE.Color(this.colorSteps[colorIndex]);
            const nextColor = new THREE.Color(this.colorSteps[nextColorIndex]);
            pixel.color.current.copy(currentColor).lerp(nextColor, fraction);
            pixel.material.color.copy(pixel.color.current);
        });

        this.floatingNumbers.forEach((obj, index) => {
            obj.mesh.position.y += obj.velocity.y;
            obj.mesh.rotation.y += obj.rotationSpeed;
            
            if (obj.mesh.position.y < -10) {
                this.scene.remove(obj.mesh);
                this.floatingNumbers.splice(index, 1);
            }
        });

        this.stars.forEach((star, index) => {
            if (star.isExplosion) {
                star.mesh.position.x += star.velocity.x;
                star.mesh.position.y += star.velocity.y;
                star.mesh.position.z += star.velocity.z;
                star.lifespan--;
                star.mesh.material.opacity = star.lifespan / 100;

                star.color.current.lerp(star.color.target, star.color.step);
                star.mesh.material.color.copy(star.color.current);
                star.mesh.material.emissive.copy(star.color.current);

                if (star.color.current.equals(star.color.target)) {
                    star.color.target = this.getRandomRainbowColor();
                }
                
                if (star.lifespan <= 0) {
                    this.scene.remove(star.mesh);
                    this.stars.splice(index, 1);
                }
            } else {
                star.mesh.position.x += star.velocity.x;
                star.mesh.position.y += star.velocity.y;
                star.mesh.position.z += star.velocity.z;
                
                star.mesh.rotation.x += star.rotationSpeed.x;
                star.mesh.rotation.y += star.rotationSpeed.y;
                star.mesh.rotation.z += star.rotationSpeed.z;

                star.color.current.lerp(star.color.target, star.color.step);
                star.mesh.material.color.copy(star.color.current);
                star.mesh.material.emissive.copy(star.color.current);

                if (star.color.current.equals(star.color.target)) {
                    star.color.target = this.getRandomRainbowColor();
                }

                const distance = new THREE.Vector3(
                    star.mesh.position.x - this.targetPosition.x,
                    star.mesh.position.y - this.targetPosition.y,
                    star.mesh.position.z - this.targetPosition.z
                ).length();

                if (!star.hitCalculator && distance < 8) {
                    star.hitCalculator = true;
                    this.createStarExplosion(star.mesh.position.clone(), star.mesh.material.color.getHex());
                    this.scene.remove(star.mesh);
                    this.stars.splice(index, 1);
                }
            }
        });

        if (this.loadingAnimation) {
            this.loadingAnimation.rotation.z += 0.05;
        }
        
        this.updateInfinityImagePosition();
        
        this.renderer.render(this.scene, this.camera);
    }

    addSpecialEffect(type) {
        const effects = {
            'equals': {
                color: 0x22c55e,
                geometry: new THREE.TextGeometry('=', {
                    font: this.font,
                    size: 4,
                    height: 1,
                }),
                emissive: true
            },
            'clear': {
                color: 0xef4444,
                geometry: new THREE.TextGeometry('C', {
                    font: this.font,
                    size: 4,
                    height: 1,
                }),
                emissive: true
            },
            'dot': {
                color: 0x3b82f6,
                geometry: new THREE.TextGeometry('.', {
                    font: this.font,
                    size: 4,
                    height: 1,
                })
            },
            'parenthesis': {
                color: 0x8b5cf6,
                geometry: new THREE.TextGeometry('()', {
                    font: this.font,
                    size: 4,
                    height: 1,
                })
            }
        };

        const effect = effects[type];
        if (!effect) return;

        const material = new THREE.MeshPhongMaterial({
            color: effect.color,
            transparent: true,
            opacity: 0.6,
            emissive: effect.emissive ? effect.color : undefined,
            emissiveIntensity: effect.emissive ? 0.5 : undefined
        });

        const mesh = new THREE.Mesh(effect.geometry, material);
        mesh.position.set(0, 15, 0);
        mesh.rotation.x = Math.random() * Math.PI;
        mesh.rotation.y = Math.random() * Math.PI;
        this.scene.add(mesh);

        this.floatingNumbers.push({
            mesh,
            velocity: new THREE.Vector3(0, -0.05, 0),
            rotationSpeed: 0.01
        });

        if (effect.emissive) {
            const glow = new THREE.PointLight(effect.color, 1, 10);
            glow.position.copy(mesh.position);
            this.scene.add(glow);
            setTimeout(() => this.scene.remove(glow), 1000);
        }
    }

    createBackgroundPattern() {
        const patternGeometry = new THREE.PlaneGeometry(2, 2);
        const pixels = [];
        
        const centerX = 0;
        const rows = 60;
        const pixelsPerRow = 30;
        const centerGap = 20;

        this.colorSteps = [
            0x0000FF,
            0x4400FF,
            0x8800FF,
            0xFF00FF,
            0xFF0088,
            0xFF0044,
            0xFF0088,
            0xFF00FF,
            0x8800FF,
            0x4400FF,
            0x0000FF
        ];
        
        for (let row = 0; row < rows; row++) {
            for (let i = 0; i < pixelsPerRow; i++) {
                [-1, 1].forEach(side => {
                    const baseColor = this.colorSteps[0];
                    const material = new THREE.MeshBasicMaterial({
                        color: baseColor,
                        transparent: true,
                        opacity: 0.8,
                    });
                    
                    const pixel = new THREE.Mesh(patternGeometry, material);
                    const scale = 1.2;
                    pixel.scale.set(scale, scale, 1);
                    
                    const offset = centerGap + i + Math.sin(row * 0.2) * 2;
                    const y = (row - rows/2) + Math.cos(i * 0.3) * 2;
                    
                    pixel.position.set(
                        centerX + (offset * side),
                        y,
                        -20
                    );
                    
                    this.scene.add(pixel);
                    
                    pixels.push({
                        mesh: pixel,
                        initialX: pixel.position.x,
                        speed: 0.2 + Math.random() * 0.1,
                        direction: side,
                        material: material,
                        color: {
                            current: new THREE.Color(baseColor),
                            target: new THREE.Color(this.colorSteps[1])
                        }
                    });
                });
            }
        }
        
        this.colorAnimationTime = 0;
        this.colorChangeSpeed = 0.002;
        
        return pixels;
    }

    createInfinityImage() {
        const imageUrl = 'static/img/daniladreemurr.webp';

        const textureLoader = new THREE.TextureLoader();
        textureLoader.load(imageUrl, (texture) => {
            const geometry = new THREE.PlaneGeometry(10, 16);
            const material = new THREE.MeshPhongMaterial({
                map: texture,
                transparent: true,
                opacity: 0.9,
                emissive: new THREE.Color(0x444444),
                emissiveMap: texture,
                emissiveIntensity: 0.7,
                side: THREE.DoubleSide
            });

            this.infinityImage = new THREE.Mesh(geometry, material);
            this.infinityImage.position.z = -6;
            this.infinityImage.position.y = 10;

            this.infinityImage.scale.set(1, 1, 1);

            this.scene.add(this.infinityImage);
        });
    }

    getInfinityPosition(t) {
        const a = 10;
        const b = 5;
        const x = a * Math.sin(t);
        const y = b * Math.sin(t) * Math.cos(t);
        return { x, y };
    }

    updateInfinityImagePosition() {
        if (!this.infinityImage) return;

        this.infinityTime += 0.002;

        const pos = this.getInfinityPosition(this.infinityTime);

        this.infinityImage.position.x = pos.x;
        this.infinityImage.position.y = pos.y + 7;

        this.infinityImage.rotation.z = Math.sin(this.infinityTime * 0.3) * 0.2;
    }

    stopLoadingAnimation() {
        if (this.loadingAnimation) {
            this.scene.remove(this.loadingAnimation);
            this.loadingAnimation = null;
        }
    }
}

let calculator3D;

document.addEventListener('DOMContentLoaded', function() {
    console.log("Инициализация калькулятора...");

    const userLogin = window.getCookieValue('user_login');
    if (!userLogin) {
        console.log("Пользователь не авторизован, перенаправление на страницу входа");
        window.location.href = '/login';
        return;
    }
    
    console.log("✅ Пользователь авторизован как:", userLogin);
    window.currentUser = userLogin;
    
    calculator3D = new Calculator3D();

    createPixelWalls();

    document.querySelectorAll('.calculator-key').forEach(key => {
        key.addEventListener('click', handleCalculatorKey);
    });

    document.getElementById('expression').addEventListener('keypress', (e) => {
        if (e.key === 'Enter') {
            calculateExpression();
        }
    });

    if (typeof loadExpressionHistory === 'function') {
        console.log("Запуск загрузки истории вычислений...");
        loadExpressionHistory();
    }
});

function handleCalculatorKey(event) {
    const key = event.target.textContent;
    const input = document.getElementById('expression');

    const keyElement = event.target;
    keyElement.classList.add('calculator-key-active');
    setTimeout(() => keyElement.classList.remove('calculator-key-active'), 150);
    
    switch(key) {
        case 'C':
            input.value = '';
            break;
        case '=':
            calculateExpression();
            break;
        default:
            input.value += key;

            if (['+', '-', '*', '/'].includes(key)) {
                calculator3D.addFloatingOperator(key);
            } else if (!isNaN(key) || key === '.') {
                calculator3D.addFloatingNumber(key);
            }
    }
}

function getAuthHeaders() {
    const headers = {'Content-Type': 'application/json'};

    const jwtToken = localStorage.getItem('jwt_token');
    if (jwtToken) {
        headers['Authorization'] = `Bearer ${jwtToken}`;
        console.log("Добавлен заголовок Authorization с JWT токеном");
    }

    const userLogin = window.getCookieValue('user_login');
    if (userLogin) {
        headers['X-User-Login'] = userLogin;
        window.currentUser = userLogin;
    } else if (window.currentUser) {
        headers['X-User-Login'] = window.currentUser;
    }
    
    return headers;
}

async function calculateExpression() {
    const input = document.getElementById('expression');
    const expression = input.value;

    if (!expression.trim()) return;

    const userLogin = window.getCookieValue('user_login');
    if (!userLogin) {
        console.log("Пользователь не авторизован");
        showError("Пожалуйста, войдите в систему для выполнения вычислений");
        setTimeout(() => {
            window.location.href = '/login';
        }, 2000);
        return;
    }

    showLoading();
    try {
        console.log("Отправка запроса с выражением:", expression);
        const headers = getAuthHeaders();
        console.log("Заголовки для запроса:", headers);

        const response = await fetch('/api/v1/calculate', {
            method: 'POST',
            headers: headers,
            body: JSON.stringify({ expression }),
            credentials: 'include'
        });

        console.log("Статус ответа от сервера:", response.status);

        const responseText = await response.text();
        console.log("Ответ сервера (текст):", responseText);

        let data;
        try {
            data = JSON.parse(responseText);
            console.log("Распарсенный ответ:", data);
        } catch (e) {
            console.error("Не удалось распарсить ответ как JSON:", e);
            throw new Error("Ошибка в формате ответа от сервера");
        }
        
        if (!response.ok) {
            if (response.status === 401) {
                console.error("Ошибка авторизации");
                showError("Ошибка авторизации. Перенаправление на страницу входа...");
                setTimeout(() => {
                    document.cookie = 'user_login=; Path=/; Expires=Thu, 01 Jan 1970 00:00:01 GMT;';
                    document.cookie = 'session_id=; Path=/; Expires=Thu, 01 Jan 1970 00:00:01 GMT;';
                    window.currentUser = null;
                    window.location.href = '/login?t=' + Date.now();
                }, 2000);
                return;
            }
            
            if (data && data.error) {
                showError(data.error);
            } else {
                showError("Ошибка на сервере: " + response.status);
            }
            hideLoading();
            return;
        }
        
        if (data.error) {
            showError(data.error);
            hideLoading();
            return;
        }

        if (data.id) {
            console.log("Получен ID выражения:", data.id);
            pollForResult(data.id);
        } else {
            throw new Error('Некорректный формат ответа: отсутствует ID');
        }
    } catch (error) {
        console.error("Ошибка при вычислении:", error);
        showError(error.message);
        hideLoading();
    }
}
function addToHistory(expression, result) {
    const historyItem = document.createElement('div');
    historyItem.className = 'history-item opacity-0 w-full mx-auto text-center';
    historyItem.innerHTML = `
        <div class="text-sm text-white/80 mb-1">${expression}</div>
        <div class="text-xl font-bold text-white">${result}</div>
    `;
    
    const historyElement = document.getElementById('history');
    if (!historyElement) return;

    if (historyElement.style.flexDirection !== 'column-reverse') {
        historyElement.style.display = 'flex';
        historyElement.style.flexDirection = 'column-reverse';
        historyElement.style.overflowY = 'auto';
    }

    historyElement.append(historyItem);
    
    gsap.to(historyItem, {
        opacity: 1,
        y: 20,
        duration: 0.5
    });

    historyElement.scrollTop = 0;
}

function showError(message) {
    const resultDiv = document.getElementById('result');
    resultDiv.classList.remove('hidden');
    resultDiv.style.background = 'rgba(255, 0, 0, 0.2)';
    document.getElementById('result-value').textContent = 'Error: ' + message;
}

function createPixelWalls() {
    const pixelContainers = document.querySelectorAll('.pixel-container');
    const rainbowColors = [
        '#FF0000', '#FF7F00', '#FFFF00', '#00FF00', '#0000FF', '#4B0082', '#9400D3'
    ];
    
    pixelContainers.forEach(container => {
        if (!container._initialized) {
            initializePixelContainer(container, rainbowColors);
        } else {
            updatePixelContainer(container, rainbowColors);
        }
    });
}

function initializePixelContainer(container, rainbowColors) {
    container._initialized = true;
    container._pixelPool = [];
    const direction = container.dataset.direction;
    const count = parseInt(container.dataset.count || 50);
    
    const colorSections = 5;
    const pixelsPerSection = Math.ceil(count / colorSections);

    for (let section = 0; section < colorSections; section++) {
        const sectionColor = rainbowColors[section % rainbowColors.length];
        
        for (let i = 0; i < pixelsPerSection; i++) {
            const pixel = document.createElement('div');
            pixel.classList.add('pixel-base');

            const containerWidth = container.offsetWidth;
            const segmentWidth = containerWidth / pixelsPerSection;
            const randomOffset = (Math.random() * segmentWidth) - (segmentWidth / 2);
            const xPos = (i * segmentWidth) + (segmentWidth / 2) + randomOffset;
            
            pixel.style.left = `${xPos}px`;
            pixel.style.backgroundColor = sectionColor;
            pixel.dataset.direction = direction;

            pixel.style.display = 'none';
            
            container.appendChild(pixel);
            container._pixelPool.push(pixel);
        }
    }

    activatePixels(container);
}

function updatePixelContainer(container, rainbowColors) {
    const colorSections = 5;
    const pixelsPerSection = Math.ceil(container._pixelPool.length / colorSections);
    
    container._pixelPool.forEach((pixel, index) => {
        const section = Math.floor(index / pixelsPerSection);
        const sectionColor = rainbowColors[section % rainbowColors.length];

        if (pixel.style.display === 'none') {
            pixel.style.backgroundColor = sectionColor;
        }
    });

    activatePixels(container);
}

function activatePixels(container) {
    const pixels = container._pixelPool;

    const visibleCount = Math.ceil(pixels.length / 3);


    function maintainPixelFlow() {
        let activePixels = 0;
        for (const p of pixels) {
            if (p.style.display === 'block') activePixels++;
        }

        const pixelsToActivate = visibleCount - activePixels;
        if (pixelsToActivate > 0) {
            const inactivePixels = [];
            for (const p of pixels) {
                if (p.style.display === 'none') inactivePixels.push(p);
                if (inactivePixels.length >= pixelsToActivate) break;
            }

            inactivePixels.forEach(pixel => startPixelAnimation(pixel));
        }

        setTimeout(maintainPixelFlow, 100);
    }

    maintainPixelFlow();
}

function startPixelAnimation(pixel) {
    pixel.style.display = 'block';

    pixel.classList.remove('pixel-animated');
    void pixel.offsetHeight;

    pixel.style.animationDelay = '0s';
    pixel.style.animationDuration = `${0.5 + Math.random() * 0.5}s`;

    pixel.classList.add('pixel-animated');

    const animationDuration = parseFloat(pixel.style.animationDuration) * 1000;
    setTimeout(() => {
        pixel.style.display = 'none';
        pixel.classList.remove('pixel-animated');
    }, animationDuration);
}

document.addEventListener('DOMContentLoaded', () => {
    createPixelWalls();

    setInterval(() => {
        document.querySelectorAll('.pixel-container').forEach(container => {
            if (container._initialized) {
                const rainbowColors = [
                    '#FF0000', '#FF7F00', '#FFFF00', '#00FF00', '#0000FF', '#4B0082', '#9400D3'
                ];
                const colorSections = 5;
                const pixelsPerSection = Math.ceil(container._pixelPool.length / colorSections);
                
                container._pixelPool.forEach((pixel, index) => {
                    if (pixel.style.display === 'none') {
                        const section = Math.floor(index / pixelsPerSection);
                        pixel.style.backgroundColor = rainbowColors[section % rainbowColors.length];
                    }
                });
            }
        });
    }, 3000);
});
function showLoading() {
    calculator3D.startLoadingAnimation();
}

function hideLoading() {
    calculator3D.stopLoadingAnimation();
}

async function pollForResult(id) {
    console.log("Polling for result for expression ID:", id);
    
    try {
        const timestamp = new Date().getTime();
        const response = await fetch(`/api/v1/expressions/${id}?_t=${timestamp}`, {
            method: 'GET',
            headers: {
                'Content-Type': 'application/json',
                'X-User-Login': window.getCookieValue('user_login') || window.currentUser
            },
            credentials: 'include'
        });
        
        console.log("Статус ответа при запросе выражения:", response.status);
        
        if (!response.ok) {
            const errorText = await response.text();
            console.error("Ошибка при получении выражения:", errorText);
            document.getElementById('error').innerHTML = `Ошибка: ${response.status} - ${errorText}`;
            hideLoading();
            return;
        }
        
        const data = await response.json();
        console.log("Получены данные выражения:", data);
        
        let expressionData;
        if (data.expression) {
            expressionData = data.expression;
            console.log("Данные выражения извлечены из поля 'expression':", expressionData);
        } else {
            expressionData = data;
            console.log("Данные выражения непосредственно в ответе:", expressionData);
        }
        
        if (expressionData.status === "PENDING" || expressionData.status === "PROCESSING") {
            console.log("Expression is still processing, will poll again in 500ms");
            setTimeout(() => pollForResult(id), 500);
        } else if (expressionData.status === "COMPLETED") {
            console.log("Expression is completed with result:", expressionData.result);
            hideLoading();
            
            let expressionText = '';
            if (expressionData.text) {
                expressionText = expressionData.text;
            } else if (expressionData.original) {
                expressionText = expressionData.original;  
            } else {
                const expressionInput = document.getElementById('expression');
                if (expressionInput) {
                    expressionText = expressionInput.value;
                }
            }
            
            console.log("Текст выражения для отображения:", expressionText);
            
            showResult(expressionData.result, expressionText);
            
            console.log("Updating history after calculation completed");
            if (typeof loadExpressionHistory === 'function') {
                loadExpressionHistory();
            }
        } else if (expressionData.status === "ERROR") {
            hideLoading();
            console.log("Expression evaluation error:", expressionData.error);
            
            const resultElement = document.getElementById('result');
            if (resultElement) {
                resultElement.classList.remove('hidden');
                resultElement.innerHTML = `<div class="text-red-500">Ошибка: ${expressionData.error || 'Неизвестная ошибка'}</div>`;
            }
            
            if (typeof loadExpressionHistory === 'function') {
                loadExpressionHistory();
            }
        }
    } catch (error) {
        hideLoading();
        console.error('Error polling for result:', error);
        
        const errorElement = document.getElementById('error');
        if (errorElement) {
            errorElement.innerHTML = `Ошибка: ${error.message}`;
        }
        
        setTimeout(() => pollForResult(id), 1000);
    }
}

function showResult(result, expression) {
    console.log(`Отображение результата: ${result} для выражения: ${expression}`);
    hideLoading();
    
    const resultElement = document.getElementById('result');
    if (resultElement) {
        resultElement.classList.remove('hidden');
        resultElement.innerHTML = `<h3 class="text-xl font-semibold mb-2">Результат:</h3>
                                   <div id="result-value" class="text-3xl font-bold">${result}</div>`;
        
        gsap.to(resultElement, {
            opacity: 1,
            y: 20,
            duration: 0.5
        });
    }
    
    if (typeof loadExpressionHistory !== 'function') {
        addToHistory(expression, result);
    }
    
    if (calculator3D) {
        calculator3D.addFloatingNumber(result);
    }
    
    console.log("Результат отображен в интерфейсе:", result);
}