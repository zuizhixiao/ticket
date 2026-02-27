// Token 存储在 cookie 中
const TokenStore = {
    KEY: 'ticket_token',
    get() {
        const m = document.cookie.match(new RegExp('(^|;\\s*)' + this.KEY + '=([^;]*)'));
        return m ? decodeURIComponent(m[2]) : '';
    },
    set(token, maxAge = 7 * 24 * 3600) {
        document.cookie = this.KEY + '=' + encodeURIComponent(token) + '; path=/; max-age=' + maxAge;
    },
    remove() {
        document.cookie = this.KEY + '=; path=/; max-age=0';
    }
};

// 认证相关 API
const AuthAPI = {
    baseUrl: '',
    getUrl(path) {
        return (this.baseUrl || '') + '/api/auth' + path;
    },
    async request(method, path, body, useToken = false) {
        const headers = { 'Content-Type': 'application/json' };
        const token = TokenStore.get();
        if (useToken && token) {
            headers['token'] = token;
        }
        const opts = { method, headers };
        if (body && method !== 'GET') opts.body = JSON.stringify(body);
        const res = await fetch(this.getUrl(path), opts);
        const json = await res.json();
        if (json.errorCode !== 0 && json.errorCode !== undefined) {
            throw new Error(json.errorMessage || '请求失败');
        }
        return json;
    },
    async getCaptcha() {
        const json = await this.request('GET', '/captcha');
        return json.data;
    },
    async login(nickname, password) {
        const json = await this.request('POST', '/login', { nickname, password });
        return json.data;
    },
    async register(nickname, captchaId, code, password) {
        const json = await this.request('POST', '/register', { nickname, captchaId, code, password });
        return json.data;
    },
    async resetPassword(nickname, captchaId, code, password) {
        await this.request('POST', '/reset-password', { nickname, captchaId, code, password });
    },
    async getUserInfo() {
        const json = await this.request('GET', '/user-info', null, true);
        return json.data;
    },
    async getUserProducts(page = 1, num = 20) {
        const url = (this.baseUrl || '') + '/api/user/products?page=' + page + '&num=' + num;
        const res = await fetch(url, {
            headers: { 'token': TokenStore.get() }
        });
        const json = await res.json();
        if (json.errorCode !== 0 && json.errorCode !== undefined) {
            throw new Error(json.errorMessage || '获取失败');
        }
        return json.data;
    }
};

// 登录成功后保存 token 和用户信息
function saveAuth(data) {
    if (data.token) TokenStore.set(data.token);
    if (data.nickname || data.avatar || data.id) {
        localStorage.setItem('ticket_user_info', JSON.stringify({
            nickname: data.nickname,
            avatar: data.avatar,
            id: data.id
        }));
    }
}

function clearAuth() {
    TokenStore.remove();
    localStorage.removeItem('ticket_user_info');
}

function getToken() {
    return TokenStore.get();
}
