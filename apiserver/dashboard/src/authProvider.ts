import { AuthProvider } from 'react-admin';

const authProvider: AuthProvider = {
    login: async ({ username, password }: { username: string; password: string }) => {
        const formData = new URLSearchParams();
        formData.append('grant_type', 'password');
        formData.append('username', username);
        formData.append('password', password);
        formData.append('client_id', 'openldap-client');
        formData.append('client_secret', 'openldap-secret');

        const request = new Request('/oauth/token', {
          method: 'POST',
          headers: new Headers({ 'Content-Type': 'application/x-www-form-urlencoded' }),
          body: formData.toString(),
        });

        try {
          const response = await fetch(request);
          if (response.status < 200 || response.status >= 300) {
            throw new Error(response.statusText);
          }
          const { access_token } = await response.json();
          localStorage.setItem('token', access_token);
          localStorage.setItem('username', username);
          return Promise.resolve();
        } catch (error) {
          throw new Error('Incorrect Username or password');
        }
    },
    logout: () => {
        localStorage.removeItem('token');
        localStorage.removeItem('username');
        return Promise.resolve();
    },
    checkError: () => Promise.resolve(),
    checkAuth: () => {
        if (localStorage.getItem('token')) {
            return Promise.resolve();
        }
        return Promise.reject();
    },
    getPermissions: () => Promise.reject('Unknown method'),
    getIdentity: () => {
        const username = localStorage.getItem('username');
        return Promise.resolve({
            id: username || 'user',
            fullName: username || 'User',
        });
    },
};

export default authProvider;
