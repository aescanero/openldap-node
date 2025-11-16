import React from 'react';
import { Admin, Resource } from 'react-admin';
import { Graph } from './graph';
import { Login, Layout } from './layout';
import { dataProvider } from './dataProvider/customDataProvider';
import authProvider from './authProvider';
import users from './users';
import groups from './groups';

const App = () => (
  <Admin
        title="OpenLDAP Node Manager"
        dataProvider={dataProvider}
        authProvider={authProvider}
        dashboard={Graph}
        loginPage={Login}
        layout={Layout}
        disableTelemetry
        requireAuth
    >
    <Resource name="users" {...users} />
    <Resource name="groups" {...groups} />
  </Admin>
);

export default App;
