import React from 'react';
import {
    List,
    Datagrid,
    TextField,
    EmailField,
    EditButton,
    DeleteButton,
    SearchInput,
    CreateButton,
    TopToolbar,
    FilterButton,
    ExportButton,
    useListContext,
} from 'react-admin';
import { Card, CardContent, Typography } from '@mui/material';

const UserListActions = () => (
    <TopToolbar>
        <FilterButton />
        <CreateButton />
        <ExportButton />
    </TopToolbar>
);

const userFilters = [
    <SearchInput key="q" source="q" alwaysOn placeholder="Search users..." />,
];

const Empty = () => (
    <Card>
        <CardContent>
            <Typography variant="h6" paragraph>
                No users yet
            </Typography>
            <Typography variant="body2">
                Create one by clicking the "Create" button
            </Typography>
        </CardContent>
    </Card>
);

export const UserList = () => (
    <List
        filters={userFilters}
        actions={<UserListActions />}
        empty={<Empty />}
        perPage={20}
        sort={{ field: 'username', order: 'ASC' }}
    >
        <Datagrid rowClick="edit" bulkActionButtons={false}>
            <TextField source="username" label="Username" />
            <TextField source="first_name" label="First Name" />
            <TextField source="last_name" label="Last Name" />
            <EmailField source="email" label="Email" />
            <TextField source="display_name" label="Display Name" />
            <TextField source="description" label="Description" />
            <EditButton />
            <DeleteButton />
        </Datagrid>
    </List>
);
