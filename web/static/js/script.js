// 电影纪念票编辑器 - 主要JavaScript逻辑
class TicketEditor {
    constructor() {
        this.canvas = document.getElementById('ticketCanvas');
        this.ctx = this.canvas.getContext('2d');
        this.apiHost = '';
        
        // 设置Canvas的安全属性，解决微信环境下的跨域问题
        this.setupCanvasSecurity();
        
        this.currentTemplate = null; // 初始化为null，等待模板加载完成后设置
        this.posterImage = null;
        this.posterServerFilename = null; // 存储海报在服务器上的文件名
        this.templateImages = {};
        this.templateUrls = {}; // 存储模板的远程URL
        
        // 新增：字体大小自定义存储
        this.customFontSizes = {
            titleFontSize: 48,
            textFontSize: 26
        };
        
        this.init();
    }
    
    // 设置Canvas安全属性
    setupCanvasSecurity() {
        // 设置Canvas的跨域属性
        this.canvas.setAttribute('crossorigin', 'anonymous');
        
        // 尝试设置Canvas的willReadFrequently属性（如果支持）
        try {
            if (this.ctx.willReadFrequently) {
                this.ctx.willReadFrequently(true);
            }
        } catch (e) {
            // 如果不支持，忽略错误
        }
        
        // 设置Canvas的imageSmoothingEnabled属性
        this.ctx.imageSmoothingEnabled = true;
        this.ctx.imageSmoothingQuality = 'high';
    }

    async init() {
        await this.loadTemplateImages();
        this.bindEvents();
        this.setupDefaultValues();
        // 注意：renderTicket() 现在在 generateTemplateUI() 中被调用，确保模板加载完成后再渲染
    }

    async loadTemplateImages() {
        try {
            //加载模板数据
            const response = await fetch(this.apiHost + '/api/getSystemTemplate', {
                method: 'GET',
            });

            if (!response.ok) {
                throw new Error('模板获取失败');
            }

            const result = await response.json();
            if(result.errorCode != 0) {
                const errorData = await response.json();
                throw new Error(errorData.message || '模板获取失败');
            }
            const templateData = result.data.list;
            
            let loadedCount = 0;
            let totalTemplates = templateData.length;
            
            // 存储模板的完整信息（包括颜色）
            this.templateData = {};
            
            templateData.forEach((template, index) => {
                const img = new Image();
                // 设置跨域属性，允许在canvas中使用
                img.crossOrigin = 'anonymous';
                img.onload = () => {
                    // 从URL中提取模板名称
                    const templateName = this.extractTemplateName(template.url);
                    this.templateImages[templateName] = img;
                    this.templateUrls[templateName] = template.url; // 存储远程URL
                    
                    // 存储模板的颜色信息
                    this.templateData[templateName] = {
                        titleColor: template.titleColor,
                        textColor: template.textColor
                    };
                    
                    console.log(`成功加载模板: ${templateName}`);
                    
                    loadedCount++;
                    if (loadedCount === totalTemplates) {
                        console.log('所有模板加载完成');
                        this.generateTemplateUI();
                    }
                };
                img.onerror = () => {
                    console.warn(`模板加载失败: ${template.url}`);
                    loadedCount++;
                    if (loadedCount === totalTemplates) {
                        console.log('所有模板加载完成');
                        this.generateTemplateUI();
                    }
                };
                img.src = template.url;
            });
        } catch (error) {
            console.error('加载模板数据失败:', error);
            // 如果加载失败，使用默认模板
            this.loadDefaultTemplate();
        }
        
    }
    
    extractTemplateName(url) {
        // 从URL中提取模板名称（去掉路径和扩展名）
        const filename = url.split('/').pop();
        return filename.replace(/\.[^/.]+$/, '');
    }
    
    loadDefaultTemplate() {
        // 默认模板加载逻辑（作为备用）
        const defaultTemplates = [
            '深蓝.png', '深灰.png', '中蓝.png', '深绿.png',
            '红色.png', '紫色.png', '银色.png',
            '浅蓝.png', '浅灰.png', '橙色.png'
        ];
        
        let loadedCount = 0;
        let totalTemplates = defaultTemplates.length;
        
        defaultTemplates.forEach(filename => {
            const img = new Image();
            // 设置跨域属性，允许在canvas中使用
            img.crossOrigin = 'anonymous';
            img.onload = () => {
                const templateName = filename.replace(/\.[^/.]+$/, '');
                // 使用远程URL，如果filename已经是完整URL则直接使用，否则拼接
                const remoteUrl = filename.startsWith('http') ? filename : `${this.apiHost}/images/template/${filename}`;
                this.templateImages[templateName] = img;
                this.templateUrls[templateName] = remoteUrl;
                
                // 设置默认颜色
                this.templateData[templateName] = {
                    titleColor: '#ffffff',
                    textColor: '#f9d9c2'
                };
                
                console.log(`成功加载默认模板: ${templateName}`);
                
                loadedCount++;
                if (loadedCount === totalTemplates) {
                    console.log('所有默认模板加载完成');
                    this.generateTemplateUI();
                    // generateTemplateUI 已经处理了模板选择和渲染，这里不需要重复设置
                }
            };
            img.onerror = () => {
                console.warn(`默认模板加载失败: ${filename}`);
                loadedCount++;
                if (loadedCount === totalTemplates) {
                    console.log('所有默认模板加载完成');
                    this.generateTemplateUI();
                    // generateTemplateUI 已经处理了模板选择和渲染，这里不需要重复设置
                }
            };
            // 直接使用远程URL，如果filename已经是完整URL则直接使用，否则拼接
            const remoteUrl = filename.startsWith('http') ? filename : `${this.apiHost}/images/template/${filename}`;
            img.src = remoteUrl;
        });
    }

    generateTemplateUI() {
        const templateGrid = document.getElementById('templateGrid');
        const templateNames = Object.keys(this.templateImages);
        
        templateGrid.innerHTML = '';
        
        templateNames.forEach((templateName, index) => {
            const templateItem = document.createElement('div');
            templateItem.className = 'template-item';
            templateItem.dataset.template = templateName;
            // 确保第一个模板默认被选中
            if (index === 0) {
                templateItem.classList.add('active');
                this.currentTemplate = templateName; // 立即设置当前模板
            }
            
            // 优先使用存储的URL，如果没有则使用Image对象的src
            const imageUrl = this.templateUrls[templateName] || this.templateImages[templateName]?.src || '';
            templateItem.innerHTML = `
                <div class="template-preview">
                    <img src="${imageUrl}" alt="${templateName}模板" class="template-image">
                </div>
            `;
            
            templateGrid.appendChild(templateItem);
        });
        
        // 确保有模板时，第一个模板被选中
        if (templateNames.length > 0) {
            this.currentTemplate = templateNames[0];
            // 强制重新渲染票根
            this.renderTicket();
        }
        
        // 重新绑定事件
        this.bindTemplateEvents();
    }

    bindTemplateEvents() {
        document.querySelectorAll('.template-item').forEach(item => {
            item.addEventListener('click', (e) => {
                this.selectTemplate(e.currentTarget.dataset.template);
            });
        });
    }

    updateTemplateUI(templateName, imageSrc) {
        // 更新模板选择器的显示
        const templateItem = document.querySelector(`[data-template="${templateName}"]`);
        if (templateItem) {
            const imgElement = templateItem.querySelector('.template-image');
            if (imgElement) {
                imgElement.src = imageSrc;
            }
        }
    }

    bindEvents() {
        // 海报上传事件
        const uploadBtn = document.getElementById('uploadBtn');
        const posterInput = document.getElementById('posterInput');

        uploadBtn.addEventListener('click', () => posterInput.click());
        posterInput.addEventListener('change', this.handleFileSelect.bind(this));

        // 移除海报事件
        document.getElementById('removePoster').addEventListener('click', () => {
            this.removePoster();
        });

        // 票根信息输入事件
        document.getElementById('movieTitle').addEventListener('input', () => this.renderTicket());
        document.getElementById('language').addEventListener('change', () => this.renderTicket());
        document.getElementById('format').addEventListener('change', () => this.renderTicket());
        document.getElementById('cinemaName').addEventListener('input', () => this.renderTicket());
        document.getElementById('hallSeat').addEventListener('input', () => this.renderTicket());
        document.getElementById('startTime').addEventListener('input', () => this.renderTicket());
        document.getElementById('exclusiveName').addEventListener('input', () => this.renderTicket());
        
        // 新增：字体大小控制事件
        this.bindFontSizeEvents();
        
        // 根据环境调整按钮文字
        this.adjustButtonText();

        // 控制按钮事件
        document.getElementById('downloadBtn').addEventListener('click', () => this.downloadTicket());
        document.getElementById('resetBtn').addEventListener('click', () => this.resetEditor());
    }
    
    // 新增：字体大小事件绑定方法
    bindFontSizeEvents() {
        // 字体大小滑块事件
        const fontSizeSliders = ['titleFontSize', 'textFontSize'];
        
        fontSizeSliders.forEach(fontType => {
            const slider = document.getElementById(fontType);
            const valueDisplay = document.getElementById(fontType + 'Value');
            
            if (slider && valueDisplay) {
                // 设置初始值
                slider.value = this.customFontSizes[fontType];
                valueDisplay.textContent = this.customFontSizes[fontType] + 'px';
                
                // 添加滑块变化事件
                slider.addEventListener('input', (e) => {
                    const newValue = parseInt(e.target.value);
                    this.customFontSizes[fontType] = newValue;
                    valueDisplay.textContent = newValue + 'px';
                    this.renderTicket(); // 实时更新预览
                });
            }
        });
        

    }
    

    
    // 根据环境调整按钮文字
    adjustButtonText() {
        const downloadBtn = document.getElementById('downloadBtn');
        if (this.isWeChatBrowser()) {
            downloadBtn.innerHTML = '💾 生成图片';
            downloadBtn.title = '请先上传海报，然后点击生成图片，长按保存到相册';
        } else {
            downloadBtn.innerHTML = '💾 下载票根';
            downloadBtn.title = '请先上传海报，然后下载票根图片到本地';
        }
    }
    
    // 更新按钮提示信息
    updateButtonTooltip() {
        const downloadBtn = document.getElementById('downloadBtn');
        if (this.posterImage) {
            // 海报已上传，更新提示信息
            if (this.isWeChatBrowser()) {
                downloadBtn.title = '点击生成图片，然后长按保存到相册';
            } else {
                downloadBtn.title = '下载票根图片到本地';
            }
        } else {
            // 海报未上传，显示需要上传的提示
            if (this.isWeChatBrowser()) {
                downloadBtn.title = '请先上传海报，然后点击生成图片，长按保存到相册';
            } else {
                downloadBtn.title = '请先上传海报，然后下载票根图片到本地';
            }
        }
    }

    selectTemplate(templateName) {
        document.querySelectorAll('.template-item').forEach(item => {
            item.classList.remove('active');
        });
        const selectedItem = document.querySelector(`[data-template="${templateName}"]`);
        if (selectedItem) {
            selectedItem.classList.add('active');
            this.currentTemplate = templateName;
            
            // 切换模板时，可以选择是否应用模板的默认字体大小
            // 这里保持用户的自定义字体大小，不自动应用模板默认值
            this.renderTicket();
        }
    }



    handleFileSelect(e) {
        const file = e.target.files[0];
        if (file) {
            // 先加载海报到前端预览，然后异步上传到服务器
            this.loadPoster(file);
            this.uploadPosterToServerAsync(file);
        }
    }

    // 异步上传海报到服务器（无提示）
    async uploadPosterToServerAsync(file) {
        try {
            // 创建FormData
            const formData = new FormData();
            formData.append('image', file);
            formData.append('type', 'poster'); // 指定为海报类型
            
            // 发送上传请求
            const response = await fetch(this.apiHost + '/api/upload/image', {
                method: 'POST',
                body: formData
            });
            
            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.message || '海报上传失败');
            }
            
            const result = await response.json();
            if (result.errcode == 0) {
                console.log('海报上传成功:', result.data);
                // 存储服务器文件名
                this.posterServerFilename = result.data.filename;
            } else {
                throw new Error(result.message || '海报上传失败');
            }
        } catch (error) {
            console.error('海报上传失败:', error);
            // 静默处理错误，不影响用户体验
        }
    }



    loadPoster(file) {
        if (!file.type.startsWith('image/')) {
            alert('请选择图片文件！');
            return;
        }

        const reader = new FileReader();
        reader.onload = (e) => {
            const img = new Image();
            img.onload = () => {
                this.posterImage = img;
                this.showPosterPreview(img.src);
                this.renderTicket();
            };
            img.src = e.target.result;
        };
        reader.readAsDataURL(file);
    }

    showPosterPreview(src) {
        document.getElementById('posterImage').src = src;
        document.getElementById('posterPreview').style.display = 'block';
        document.getElementById('uploadBtn').style.display = 'none';
        
        this.renderTicket();
        
        // 更新按钮提示信息
        this.updateButtonTooltip();
    }

    removePoster() {
        this.posterImage = null;
        this.posterServerFilename = null; // 清除服务器文件名
        document.getElementById('posterPreview').style.display = 'none';
        document.getElementById('uploadBtn').style.display = 'inline-block';
        document.getElementById('posterInput').value = '';
        this.renderTicket();
        
        // 更新按钮提示信息
        this.updateButtonTooltip();
    }



    setupDefaultValues() {
        // 不再自动设置开始时间，让用户自己输入
        // 如果需要默认时间，可以在这里设置，但当前选择不设置
    }

    renderTicket() {
        const canvas = this.canvas;
        const ctx = this.ctx;
        
        ctx.clearRect(0, 0, canvas.width, canvas.height);
        
        this.drawBackground();
        
        if (this.posterImage) {
            this.drawPoster();
        }
        
        this.drawTicketInfo();
        this.drawDecorations();
    }

    drawBackground() {
        const ctx = this.ctx;
        const canvas = this.canvas;
        
        // 绘制模板背景图片
        if (this.templateImages[this.currentTemplate]) {
            ctx.drawImage(this.templateImages[this.currentTemplate], 0, 0, canvas.width, canvas.height);
        } else {
            // 如果图片未加载，使用默认背景
            ctx.fillStyle = '#ffffff';
            ctx.fillRect(0, 0, canvas.width, canvas.height);
            ctx.strokeStyle = '#e9ecef';
            ctx.lineWidth = 2;
            ctx.strokeRect(10, 10, canvas.width - 20, canvas.height - 20);
        }
    }



    drawPoster() {
        if (!this.posterImage) return;
        
        const ctx = this.ctx;
        const canvas = this.canvas;
        
        const posterWidth = 617;
        const posterHeight = 845;
        const posterX = (canvas.width - posterWidth) / 2;
        const posterY = 40;
        const borderRadius = 20; // 圆角半径
        
        // 创建圆角矩形路径
        ctx.save();
        ctx.beginPath();
        ctx.moveTo(posterX + borderRadius, posterY);
        ctx.lineTo(posterX + posterWidth - borderRadius, posterY);
        ctx.quadraticCurveTo(posterX + posterWidth, posterY, posterX + posterWidth, posterY + borderRadius);
        ctx.lineTo(posterX + posterWidth, posterY + posterHeight - borderRadius);
        ctx.quadraticCurveTo(posterX + posterWidth, posterY + posterHeight, posterX + posterWidth - borderRadius, posterY + posterHeight);
        ctx.lineTo(posterX + borderRadius, posterY + posterHeight);
        ctx.quadraticCurveTo(posterX, posterY + posterHeight, posterX, posterY + posterHeight - borderRadius);
        ctx.lineTo(posterX, posterY + borderRadius);
        ctx.quadraticCurveTo(posterX, posterY, posterX + borderRadius, posterY);
        ctx.closePath();
        
        // 应用阴影
        ctx.shadowColor = 'rgba(0,0,0,0.3)';
        ctx.shadowBlur = 15;
        ctx.shadowOffsetX = 8;
        ctx.shadowOffsetY = 8;
        
        // 绘制海报图片（应用裁剪路径）
        ctx.clip();
        ctx.drawImage(this.posterImage, posterX, posterY, posterWidth, posterHeight);
        
        // 重置阴影
        ctx.shadowColor = 'transparent';
        ctx.shadowBlur = 0;
        ctx.shadowOffsetX = 0;
        ctx.shadowOffsetY = 0;
        
        // 绘制圆角边框
        ctx.restore();
        ctx.save();
        ctx.beginPath();
        ctx.moveTo(posterX + borderRadius, posterY);
        ctx.lineTo(posterX + posterWidth - borderRadius, posterY);
        ctx.quadraticCurveTo(posterX + posterWidth, posterY, posterX + posterWidth, posterY + borderRadius);
        ctx.lineTo(posterX + posterWidth, posterY + posterHeight - borderRadius);
        ctx.quadraticCurveTo(posterX + posterWidth, posterY + posterHeight, posterX + posterWidth - borderRadius, posterY + posterHeight);
        ctx.lineTo(posterX + borderRadius, posterY + posterHeight);
        ctx.quadraticCurveTo(posterX, posterY + posterHeight, posterX, posterY + posterHeight - borderRadius);
        ctx.lineTo(posterX, posterY + borderRadius);
        ctx.quadraticCurveTo(posterX, posterY, posterX + borderRadius, posterY);
        ctx.closePath();
        
        ctx.strokeStyle = '#ffffff';
        ctx.lineWidth = 4;
        ctx.stroke();
        ctx.restore();
    }

    drawTicketInfo() {
        const ctx = this.ctx;
        const canvas = this.canvas;
        
        const movieTitle = document.getElementById('movieTitle').value || '电影标题';
        const cinemaName = document.getElementById('cinemaName').value || '影院名称';
        const hallSeat = document.getElementById('hallSeat').value || '影厅号及座位号';
        const startTime = document.getElementById('startTime').value || '';
        
        let timeText = startTime || '开始时间';
        
        ctx.textAlign = 'center';
        
        // 获取模板的默认颜色
        let textColor = '#f9d9c2'; // 默认文本颜色
        let titleColor = '#ffffff'; // 默认电影标题颜色为白色
        
        // 使用自定义字体大小
        let titleFontSize = this.customFontSizes.titleFontSize;
        let textFontSize = this.customFontSizes.textFontSize;
        const nameFontSize = 35; // 特殊ID字体大小（固定）
        const tailFontSize = 24; // 底部文字字体大小（固定）
        
        // 使用当前模板的颜色配置
        if (this.templateData && this.templateData[this.currentTemplate]) {
            textColor = this.templateData[this.currentTemplate].textColor;
            titleColor = this.templateData[this.currentTemplate].titleColor;
        }
        
        // 调整文字位置，改为左对齐
        ctx.textAlign = 'left';
        const textStartX = 34;  // 文字起始位置
        
        // 绘制电影标题（使用单独的颜色）
        const firstLineY = 1017; //电影标题Y值 固定
        const tailLineY = 1306;  //专属名称Y值 固定
        const firstBottomMargin = 35; //电影标题与语言标签的间距 固定
        const textBottomMargin = 17; //语言标签与影厅号及座位号的间距 固定
        ctx.font = `bold ${titleFontSize}px Microsoft YaHei`;
        ctx.fillStyle = titleColor;
        ctx.fillText(movieTitle, textStartX, firstLineY);
        
        // 绘制语言标签、格式标签和影院名称（在同一行）
        const language = document.getElementById('language').value || '国语';
        let format = document.getElementById('format').value || '2D';
        let secondText = language + " " + format + " | " + cinemaName;
        const fontStyle = `${textFontSize}px Microsoft YaHei`;

        // 绘制语言标签
        const secondLineY = firstLineY + titleFontSize + firstBottomMargin;
        ctx.font = fontStyle;
        ctx.fillStyle = textColor;
        ctx.fillText(secondText, textStartX, secondLineY);
        
        // 绘制其他信息
        ctx.fillText(hallSeat, textStartX, secondLineY + textFontSize + textBottomMargin);
        ctx.fillText(timeText, textStartX, secondLineY + (textFontSize + textBottomMargin)*2);
        
        // 绘制专属名称（在最下面居中显示）
        const nameBottomMargin = 7;
        ctx.font = `${nameFontSize}px Microsoft YaHei`;
        const exclusiveName = document.getElementById('exclusiveName').value || '特殊ID';
        ctx.textAlign = 'center'; // 居中对齐
        ctx.fillStyle = titleColor; // 使用与文本相同的颜色
        ctx.fillText(`@${exclusiveName}`, canvas.width / 2, tailLineY); // 在票根最下面居中显示，前面加上@符号  
        ctx.font = `${tailFontSize}px Microsoft YaHei`;
        ctx.fillStyle = textColor; // 使用与文本相同的颜色
        ctx.fillText("专属纪念票", canvas.width / 2, tailLineY + nameBottomMargin + nameFontSize)

    }

    drawDecorations() {
        const ctx = this.ctx;
        const canvas = this.canvas;
        
        // 根据模板添加适当的装饰
        if (!this.currentTemplate) {
            // 如果没有选中模板，不绘制装饰
            return;
        }
        
        switch (this.currentTemplate) {
            case 'classic':
                this.drawCornerDecoration('#007bff', 20, 20);
                this.drawCornerDecoration('#007bff', canvas.width - 20, 20);
                this.drawCornerDecoration('#007bff', 20, canvas.height - 20);
                this.drawCornerDecoration('#007bff', canvas.width - 20, canvas.height - 20);
                break;
                
            case 'modern':
                this.drawGeometricDecoration();
                break;
                
            case 'vintage':
                this.drawVintageDecoration();
                break;
                
            case 'elegant':
                this.drawElegantDecoration();
                break;
                
            default:
                // 对于其他模板，可以添加默认装饰或保持空白
                break;
        }
    }

    drawCornerDecoration(color, x, y) {
        const ctx = this.ctx;
        const size = 15;
        
        ctx.save();
        ctx.translate(x, y);
        ctx.strokeStyle = color;
        ctx.lineWidth = 2;
        
        ctx.beginPath();
        ctx.moveTo(-size/2, -size/2);
        ctx.lineTo(size/2, -size/2);
        ctx.moveTo(-size/2, -size/2);
        ctx.lineTo(-size/2, size/2);
        ctx.stroke();
        
        ctx.restore();
    }

    drawGeometricDecoration() {
        const ctx = this.ctx;
        const canvas = this.canvas;
        
        ctx.save();
        ctx.strokeStyle = 'rgba(255,255,255,0.3)';
        ctx.lineWidth = 2;
        
        for (let i = 0; i < 3; i++) {
            const x = 80 + i * 80;
            const y = 900;  // 向上移动装饰元素
            const size = 40;
            
            ctx.beginPath();
            ctx.arc(x, y, size, 0, Math.PI * 2);
            ctx.stroke();
        }
        
        ctx.restore();
    }

    drawVintageDecoration() {
        const ctx = this.ctx;
        const canvas = this.canvas;
        
        ctx.save();
        ctx.strokeStyle = '#8b4513';
        ctx.lineWidth = 2;
        ctx.globalAlpha = 0.6;
        
        for (let i = 0; i < 5; i++) {
            const x = 80 + i * 120;
            const y = 900;  // 向上移动装饰元素
            this.drawVintagePattern(x, y, 25);
        }
        
        ctx.restore();
    }

    drawVintagePattern(x, y, size) {
        const ctx = this.ctx;
        
        ctx.beginPath();
        ctx.moveTo(x, y - size);
        ctx.lineTo(x + size/2, y);
        ctx.lineTo(x, y + size);
        ctx.lineTo(x - size/2, y);
        ctx.closePath();
        ctx.stroke();
    }

    drawElegantDecoration() {
        const ctx = this.ctx;
        const canvas = this.canvas;
        
        ctx.save();
        ctx.strokeStyle = 'rgba(255,255,255,0.4)';
        ctx.lineWidth = 3;
        
        ctx.beginPath();
        ctx.moveTo(80, 900);  // 向上移动装饰元素
        ctx.quadraticCurveTo(canvas.width/2, 880, canvas.width - 80, 900);
        ctx.stroke();
        
        ctx.restore();
    }

    async downloadTicket() {
        try {
            // 校验海报是否上传
            if (!this.posterImage) {
                this.showPosterRequiredTip();
                return;
            }
            
            // 校验必填字段
            const movieTitle = document.getElementById('movieTitle').value.trim();
            const hallSeat = document.getElementById('hallSeat').value.trim();
            const startTime = document.getElementById('startTime').value.trim();
            const exclusiveName = document.getElementById('exclusiveName').value.trim();
            
            if (!movieTitle) {
                this.showFieldRequiredTip('电影标题');
                return;
            }
            
            if (!hallSeat) {
                this.showFieldRequiredTip('影厅号及座位号');
                return;
            }
            
            if (!startTime) {
                this.showFieldRequiredTip('开始时间');
                return;
            }
            
            if (!exclusiveName) {
                this.showFieldRequiredTip('特殊ID');
                return;
            }
            
            const canvas = this.canvas;
            
            // 获取电影标题作为文件名的一部分
            const timestamp = new Date().getTime();
            const cleanTitle = movieTitle.replace(/[<>:"/\\|?*]/g, '_');
            
            // 先上传图片到服务端
            try {
                await this.uploadImageToServer(canvas, cleanTitle, timestamp);
            } catch (uploadError) {
                console.warn('图片上传失败，继续本地下载:', uploadError);
                // 上传失败不影响本地下载功能
            }
            
            // 检测是否在微信环境中
            const isWeChat = this.isWeChatBrowser();
            
            if (isWeChat) {
                // 微信环境下，使用长按保存的方式
                this.saveForWeChat(canvas, cleanTitle, timestamp);
            } else {
                // 非微信环境，使用常规下载方式
                this.downloadForNormalBrowser(canvas, cleanTitle, timestamp);
            }
            
        } catch (error) {
            console.error('保存失败:', error);
            alert('保存失败，请重试');
        }
    }
    
    // 显示字段必填提示
    showFieldRequiredTip(fieldName) {
        const tipBox = document.createElement('div');
        tipBox.style.cssText = `
            position: fixed;
            top: 50%;
            left: 50%;
            transform: translate(-50%, -50%);
            background: white;
            padding: 25px;
            border-radius: 15px;
            box-shadow: 0 8px 30px rgba(0,0,0,0.3);
            z-index: 10001;
            text-align: center;
            max-width: 450px;
            border: 2px solid #dc3545;
        `;
        
        tipBox.innerHTML = `
            <div style="margin-bottom: 20px;">
                <span style="font-size: 48px;">⚠️</span>
            </div>
            <h3 style="margin: 0 0 20px 0; color: #dc3545; font-size: 20px;">请填写${fieldName}</h3>
            <div style="display: flex; gap: 15px; justify-content: center; width: 100%;">
                <button id="closeFieldTip" style="
                    padding: 15px 20px;
                    background: linear-gradient(135deg, #6c757d, #495057);
                    color: white;
                    border: none;
                    border-radius: 25px;
                    cursor: pointer;
                    font-size: 16px;
                    font-weight: bold;
                    box-shadow: 0 4px 15px rgba(108,117,125,0.3);
                    transition: all 0.3s ease;
                    flex: 1;
                    min-width: 0;
                    white-space: nowrap;
                " onmouseover="this.style.transform='scale(1.02)'" onmouseout="this.style.transform='scale(1)'">我知道了</button>
            </div>
        `;
        
        document.body.appendChild(tipBox);
        
        // 添加关闭事件
        document.getElementById('closeFieldTip').addEventListener('click', () => {
            document.body.removeChild(tipBox);
        });
        
        // 5秒后自动关闭
        setTimeout(() => {
            if (document.body.contains(tipBox)) {
                document.body.removeChild(tipBox);
            }
        }, 5000);
    }
    
    // 显示海报必填提示
    showPosterRequiredTip() {
        const tipBox = document.createElement('div');
        tipBox.style.cssText = `
            position: fixed;
            top: 50%;
            left: 50%;
            transform: translate(-50%, -50%);
            background: white;
            padding: 25px;
            border-radius: 15px;
            box-shadow: 0 8px 30px rgba(0,0,0,0.3);
            z-index: 10001;
            text-align: center;
            max-width: 450px;
            border: 2px solid #ffc107;
        `;
        
        tipBox.innerHTML = `
            <div style="margin-bottom: 20px;">
                <span style="font-size: 48px;">🎬</span>
            </div>
            <h3 style="margin: 0 0 20px 0; color: #ffc107; font-size: 20px;">请先上传海报</h3>
            <div style="display: flex; gap: 15px; justify-content: center; width: 100%;">
                <button id="uploadPosterNow" style="
                    padding: 15px 20px;
                    background: linear-gradient(135deg, #28a745, #20c997);
                    color: white;
                    border: none;
                    border-radius: 25px;
                    cursor: pointer;
                    font-size: 16px;
                    font-weight: bold;
                    box-shadow: 0 4px 15px rgba(40,167,69,0.3);
                    transition: all 0.3s ease;
                    flex: 1;
                    min-width: 0;
                    white-space: nowrap;
                " onmouseover="this.style.transform='scale(1.02)'" onmouseout="this.style.transform='scale(1)'">立即上传</button>
                <button id="closePosterTip" style="
                    padding: 15px 20px;
                    background: linear-gradient(135deg, #6c757d, #495057);
                    color: white;
                    border: none;
                    border-radius: 25px;
                    cursor: pointer;
                    font-size: 16px;
                    font-weight: bold;
                    box-shadow: 0 4px 15px rgba(108,117,125,0.3);
                    transition: all 0.3s ease;
                    flex: 1;
                    min-width: 0;
                    white-space: nowrap;
                " onmouseover="this.style.transform='scale(1.02)'" onmouseout="this.style.transform='scale(1)'">稍后再说</button>
            </div>
        `;
        
        document.body.appendChild(tipBox);
        
        // 添加关闭事件
        document.getElementById('closePosterTip').addEventListener('click', () => {
            document.body.removeChild(tipBox);
        });
        
        // 添加立即上传事件
        document.getElementById('uploadPosterNow').addEventListener('click', () => {
            document.body.removeChild(tipBox);
            // 触发海报上传
            document.getElementById('uploadBtn').click();
        });
        
        // 8秒后自动关闭（给用户更多时间阅读和操作）
        setTimeout(() => {
            if (document.body.contains(tipBox)) {
                document.body.removeChild(tipBox);
            }
        }, 8000);
    }
    
    // 检测是否在微信浏览器中
    isWeChatBrowser() {
        const ua = navigator.userAgent.toLowerCase();
        return ua.includes('micromessenger');
    }
    
    // 微信环境下的保存方式
    saveForWeChat(canvas, cleanTitle, timestamp) {
        try {
            // 创建临时图片元素
            const tempImg = document.createElement('img');
            tempImg.style.cssText = `
                position: fixed;
                top: 0;
                left: 0;
                width: 100%;
                height: 100%;
                z-index: 9999;
                background: rgba(0,0,0,0.8);
                object-fit: contain;
                cursor: pointer;
            `;
            
            // 添加提示文字
            const tip = document.createElement('div');
            tip.style.cssText = `
                position: fixed;
                top: 20px;
                left: 50%;
                transform: translateX(-50%);
                color: white;
                font-size: 20px;
                font-weight: bold;
                z-index: 10000;
                text-align: center;
                background: linear-gradient(135deg, #007bff, #0056b3);
                padding: 15px 30px;
                border-radius: 25px;
                box-shadow: 0 4px 20px rgba(0,123,255,0.4);
                border: 2px solid rgba(255,255,255,0.3);
                animation: pulse 2s infinite;
            `;
            tip.innerHTML = `
                <div style="display: flex; align-items: center; justify-content: center; gap: 10px;">
                    <span style="font-size: 24px;">💾</span>
                    <div>
                        <div style="font-size: 18px; margin-bottom: 5px;">长按图片保存到相册</div>
                        <div style="font-size: 14px; opacity: 0.9;">点击图片关闭</div>
                    </div>
                </div>
            `;
            
            // 添加脉冲动画样式
            const style = document.createElement('style');
            style.textContent = `
                @keyframes pulse {
                    0% { transform: translateX(-50%) scale(1); }
                    50% { transform: translateX(-50%) scale(1.05); }
                    100% { transform: translateX(-50%) scale(1); }
                }
            `;
            document.head.appendChild(style);
            
            // 尝试多种方式生成图片
            let imageGenerated = false;
            
            // 方法1: 使用toBlob
            try {
                canvas.toBlob((blob) => {
                    if (blob && !imageGenerated) {
                        imageGenerated = true;
                        const url = URL.createObjectURL(blob);
                        tempImg.src = url;
                        this.setupImageDisplay(tempImg, tip, url);
                    }
                }, 'image/png', 0.9);
            } catch (e) {
                console.warn('toBlob方法失败:', e);
            }
            
            // 方法2: 使用toDataURL作为备选
            setTimeout(() => {
                if (!imageGenerated) {
                    try {
                        const dataURL = canvas.toDataURL('image/png', 0.9);
                        tempImg.src = dataURL;
                        this.setupImageDisplay(tempImg, tip);
                        imageGenerated = true;
                    } catch (e) {
                        console.warn('toDataURL方法也失败:', e);
                        // 如果两种方法都失败，显示错误提示
                        this.showWeChatError();
                    }
                }
            }, 100);
            
        } catch (error) {
            console.error('微信保存失败:', error);
            this.showWeChatError();
        }
    }
    
    // 设置图片显示
    setupImageDisplay(tempImg, tip, url = null) {
        // 添加点击关闭事件
        tempImg.addEventListener('click', () => {
            document.body.removeChild(tempImg);
            document.body.removeChild(tip);
            if (url) {
                URL.revokeObjectURL(url);
            }
        });
        
        // 添加到页面
        document.body.appendChild(tip);
        document.body.appendChild(tempImg);
        
        // 显示成功提示
        this.showWeChatTip();
    }
    
    // 显示微信环境下的错误提示
    showWeChatError() {
        const errorBox = document.createElement('div');
        errorBox.style.cssText = `
            position: fixed;
            top: 50%;
            left: 50%;
            transform: translate(-50%, -50%);
            background: white;
            padding: 30px;
            border-radius: 20px;
            box-shadow: 0 10px 40px rgba(220,53,69,0.2);
            z-index: 10001;
            text-align: center;
            max-width: 400px;
            border: 3px solid #dc3545;
            animation: slideIn 0.3s ease-out;
        `;
        
        errorBox.innerHTML = `
             <div style="margin-bottom: 25px;">
                 <span style="font-size: 64px;">❌</span>
             </div>
             <h3 style="margin: 0 0 20px 0; color: #dc3545; font-size: 24px; font-weight: bold;">保存失败</h3>
             <div style="display: flex; gap: 15px; justify-content: center;">
                <button id="tryScreenshot" style="
                    padding: 14px 28px;
                    background: linear-gradient(135deg, #28a745, #20c997);
                    color: white;
                    border: none;
                    border-radius: 25px;
                    cursor: pointer;
                    font-size: 16px;
                    font-weight: bold;
                    box-shadow: 0 4px 15px rgba(40,167,69,0.3);
                    transition: all 0.3s ease;
                    flex: 1;
                    max-width: 160px;
                " onmouseover="this.style.transform='scale(1.05)'" onmouseout="this.style.transform='scale(1)'">尝试截图保存</button>
                <button id="closeError" style="
                    padding: 14px 28px;
                    background: linear-gradient(135deg, #dc3545, #c82333);
                    color: white;
                    border: none;
                    border-radius: 25px;
                    cursor: pointer;
                    font-size: 16px;
                    font-weight: bold;
                    box-shadow: 0 4px 15px rgba(220,53,69,0.3);
                    transition: all 0.3s ease;
                    flex: 1;
                    max-width: 160px;
                " onmouseover="this.style.transform='scale(1.05)'" onmouseout="this.style.transform='scale(1)'">知道了</button>
            </div>
        `;
        
        document.body.appendChild(errorBox);
        
        // 添加动画样式
        const style = document.createElement('style');
        style.textContent = `
            @keyframes slideIn {
                0% { 
                    opacity: 0; 
                    transform: translate(-50%, -50%) scale(0.8); 
                }
                100% { 
                    opacity: 1; 
                    transform: translate(-50%, -50%) scale(1); 
                }
            }
        `;
        document.head.appendChild(style);
        
        // 添加关闭事件
        document.getElementById('closeError').addEventListener('click', () => {
            document.body.removeChild(errorBox);
        });
        
        // 添加截图保存尝试事件
        document.getElementById('tryScreenshot').addEventListener('click', () => {
            document.body.removeChild(errorBox);
            this.tryScreenshotSave();
        });
        
        // 8秒后自动关闭（给用户更多时间阅读）
        setTimeout(() => {
            if (document.body.contains(errorBox)) {
                document.body.removeChild(errorBox);
            }
        }, 8000);
    }
    
    // 尝试截图保存方式
    tryScreenshotSave() {
        // 显示截图保存指导
        const guideBox = document.createElement('div');
        guideBox.style.cssText = `
            position: fixed;
            top: 0;
            left: 0;
            width: 100%;
            height: 100%;
            background: linear-gradient(135deg, rgba(0,0,0,0.95), rgba(0,0,0,0.85));
            z-index: 10002;
            display: flex;
            flex-direction: column;
            align-items: center;
            justify-content: center;
            color: white;
            text-align: center;
            padding: 30px;
            animation: fadeIn 0.4s ease-out;
        `;
        
        guideBox.innerHTML = `
            <div style="margin-bottom: 40px;">
                <span style="font-size: 80px;">📱</span>
            </div>
            <h2 style="margin-bottom: 40px; color: #ffd700; font-size: 32px; font-weight: bold; text-shadow: 0 2px 10px rgba(255,215,0,0.3);">截图保存指导</h2>
            <div style="max-width: 500px; line-height: 1.8; margin-bottom: 40px;">
                <div style="background: rgba(255,255,255,0.1); padding: 25px; border-radius: 20px; margin-bottom: 25px; border: 1px solid rgba(255,255,255,0.2); backdrop-filter: blur(10px);">
                    <h3 style="margin: 0 0 20px 0; color: #ffd700; font-size: 20px;">微信截图</h3>
                    <p style="color: #fff; font-size: 16px;">同时按下音量下键 + 电源键进行截图</p>
                </div>
                
                <div style="background: rgba(255,255,255,0.1); padding: 25px; border-radius: 20px; margin-bottom: 25px; border: 1px solid rgba(255,255,255,0.2); backdrop-filter: blur(10px);">
                    <h3 style="margin: 0 0 20px 0; color: #ffd700; font-size: 20px;">分享保存</h3>
                    <p style="color: #fff; font-size: 16px;">截图后分享给文件传输助手，长按保存</p>
                </div>
                
                <div style="background: rgba(255,255,255,0.1); padding: 25px; border-radius: 20px; border: 1px solid rgba(255,255,255,0.2); backdrop-filter: blur(10px);">
                    <h3 style="margin: 0 0 20px 0; color: #ffd700; font-size: 20px;">浏览器打开</h3>
                    <p style="color: #fff; font-size: 16px;">在浏览器中打开页面，使用保存功能</p>
                </div>
            </div>
            <button id="closeGuide" style="
                padding: 16px 32px;
                background: linear-gradient(135deg, #007bff, #0056b3);
                color: white;
                border: none;
                border-radius: 30px;
                font-size: 18px;
                font-weight: bold;
                cursor: pointer;
                box-shadow: 0 6px 20px rgba(0,123,255,0.4);
                transition: all 0.3s ease;
                border: 2px solid rgba(255,255,255,0.3);
            " onmouseover="this.style.transform='scale(1.05)'" onmouseout="this.style.transform='scale(1)'">我明白了</button>
        `;
        
        document.body.appendChild(guideBox);
        
        // 添加动画样式
        const style = document.createElement('style');
        style.textContent = `
            @keyframes fadeIn {
                0% { 
                    opacity: 0; 
                    transform: scale(0.9); 
                }
                100% { 
                    opacity: 1; 
                    transform: scale(1); 
                }
            }
        `;
        document.head.appendChild(style);
        
        // 添加关闭事件
        document.getElementById('closeGuide').addEventListener('click', () => {
            document.body.removeChild(guideBox);
        });
        
        // 10秒后自动关闭（给用户充足时间阅读）
        setTimeout(() => {
            if (document.body.contains(guideBox)) {
                document.body.removeChild(guideBox);
            }
        }, 10000);
    }
    
    // 上传图片到服务端
    async uploadImageToServer(canvas, cleanTitle, timestamp) {
        return new Promise((resolve, reject) => {
            // 将canvas转换为blob
            canvas.toBlob(async (blob) => {
                if (!blob) {
                    reject(new Error('Canvas转换为Blob失败'));
                    return;
                }

                // 创建FormData
                const formData = new FormData();
                formData.append('image', blob, `${cleanTitle}_${timestamp}.png`);
                formData.append('type', 'product'); // 指定为成品类型

                try {
                    // 发送上传请求
                    const response = await fetch(this.apiHost + '/api/upload/image', {
                        method: 'POST',
                        body: formData
                    });

                    if (!response.ok) {
                        const errorData = await response.json();
                        throw new Error(errorData.message || '上传失败');
                    }

                    const result = await response.json();
                    if (result.errcode == 0) {
                        console.log('图片上传成功:', result.data);
                        // 显示上传成功提示
                        // this.showUploadSuccessTip(result.data.filename);
                        resolve(result);
                    } else {
                        throw new Error(result.message || '上传失败');
                    }
                } catch (error) {
                    console.error('图片上传失败:', error);
                    reject(error);
                }
            }, 'image/png', 0.9);
        });
    }

    // 显示上传成功提示
    showUploadSuccessTip(filename) {
        const tipBox = document.createElement('div');
        tipBox.style.cssText = `
            position: fixed;
            top: 20px;
            right: 20px;
            background: linear-gradient(135deg, #28a745, #20c997);
            color: white;
            padding: 15px 20px;
            border-radius: 25px;
            box-shadow: 0 4px 15px rgba(40,167,69,0.3);
            z-index: 10001;
            font-size: 14px;
            font-weight: bold;
            border: 2px solid rgba(255,255,255,0.3);
            animation: slideInRight 0.3s ease-out;
        `;
        
        tipBox.innerHTML = `
            <div style="display: flex; align-items: center; gap: 8px;">
                <span style="font-size: 18px;">✅</span>
                <span>图片已上传至服务端</span>
            </div>
        `;
        
        // 添加动画样式
        const style = document.createElement('style');
        style.textContent = `
            @keyframes slideInRight {
                0% { 
                    opacity: 0; 
                    transform: translateX(100%); 
                }
                100% { 
                    opacity: 1; 
                    transform: translateX(0); 
                }
            }
        `;
        document.head.appendChild(style);
        
        document.body.appendChild(tipBox);
        
        // 3秒后自动关闭
        setTimeout(() => {
            if (document.body.contains(tipBox)) {
                document.body.removeChild(tipBox);
            }
        }, 3000);
    }

    // 常规浏览器的下载方式
    downloadForNormalBrowser(canvas, cleanTitle, timestamp) {
        const link = document.createElement('a');
        link.download = `${cleanTitle}_${timestamp}.png`;
        
        canvas.toBlob((blob) => {
            if (blob) {
                const url = URL.createObjectURL(blob);
                link.href = url;
                link.click();
                setTimeout(() => URL.revokeObjectURL(url), 100);
            } else {
                alert('下载失败，请重试');
            }
        }, 'image/png');
    }
    
    // 显示微信环境下的提示
    showWeChatTip() {
        // 创建提示框
        const tipBox = document.createElement('div');
        tipBox.style.cssText = `
            position: fixed;
            top: 50%;
            left: 50%;
            transform: translate(-50%, -50%);
            background: white;
            padding: 25px;
            border-radius: 15px;
            box-shadow: 0 8px 30px rgba(0,0,0,0.3);
            z-index: 10001;
            text-align: center;
            max-width: 350px;
            border: 2px solid #007bff;
        `;
        
        tipBox.innerHTML = `
            <div style="margin-bottom: 20px;">
                <span style="font-size: 48px;">💾</span>
            </div>
            <h3 style="margin: 0 0 20px 0; color: #007bff; font-size: 20px;">图片已生成！</h3>
            <div style="background: #f8f9fa; padding: 15px; border-radius: 10px; margin-bottom: 20px;">
                <p style="margin: 0; color: #495057; line-height: 1.6; font-size: 16px;">
                    长按图片保存到相册
                </p>
            </div>
            <button id="closeTip" style="
                padding: 12px 30px;
                background: linear-gradient(135deg, #007bff, #0056b3);
                color: white;
                border: none;
                border-radius: 25px;
                cursor: pointer;
                font-size: 16px;
                font-weight: bold;
                box-shadow: 0 4px 15px rgba(0,123,255,0.3);
                transition: all 0.3s ease;
            " onmouseover="this.style.transform='scale(1.05)'" onmouseout="this.style.transform='scale(1)'">我明白了</button>
        `;
        
        document.body.appendChild(tipBox);
        
        // 添加关闭事件
        document.getElementById('closeTip').addEventListener('click', () => {
            document.body.removeChild(tipBox);
        });
        
        // 5秒后自动关闭（给用户更多时间阅读）
        setTimeout(() => {
            if (document.body.contains(tipBox)) {
                document.body.removeChild(tipBox);
            }
        }, 5000);
    }

    resetEditor() {
        // 清空所有文本输入框
        document.getElementById('movieTitle').value = '';
        document.getElementById('cinemaName').value = '';
        document.getElementById('hallSeat').value = '';
        document.getElementById('exclusiveName').value = '';
        
        // 重置下拉选择框为默认值
        document.getElementById('language').value = '国语';
        document.getElementById('format').value = '2D';
        
        // 设置默认时间
        this.setupDefaultValues();
        

        
        // 重置为第一个可用的模板
        const templateNames = Object.keys(this.templateImages);
        if (templateNames.length > 0) {
            this.selectTemplate(templateNames[0]);
        }
        this.removePoster();
        
        this.renderTicket();
    }
}

// 页面加载完成后初始化编辑器
document.addEventListener('DOMContentLoaded', async () => {
    new TicketEditor();
});
