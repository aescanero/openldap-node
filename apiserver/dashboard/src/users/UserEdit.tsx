import React from 'react';
import {
    Edit,
    SimpleForm,
    TextInput,
    email,
    required,
    minLength,
    DeleteButton,
    TopToolbar,
    ListButton,
} from 'react-admin';
import { Box, Typography } from '@mui/material';

const UserEditActions = () => (
    <TopToolbar>
        <ListButton />
        <DeleteButton />
    </TopToolbar>
);

export const UserEdit = () => (
    <Edit actions={<UserEditActions />}>
        <SimpleForm>
            <Typography variant="h6" gutterBottom>
                Edit User
            </Typography>
            <TextInput
                source="username"
                label="Username"
                disabled
                fullWidth
            />
            <Box display="flex" gap={2} width="100%">
                <TextInput
                    source="first_name"
                    label="First Name"
                    validate={required()}
                    fullWidth
                />
                <TextInput
                    source="last_name"
                    label="Last Name"
                    validate={required()}
                    fullWidth
                />
            </Box>
            <TextInput
                source="email"
                label="Email"
                type="email"
                validate={[email()]}
                fullWidth
            />
            <TextInput
                source="display_name"
                label="Display Name"
                fullWidth
            />
            <TextInput
                source="description"
                label="Description"
                multiline
                rows={3}
                fullWidth
            />
            <Typography variant="subtitle2" color="textSecondary" sx={{ mt: 2 }}>
                Change Password (leave empty to keep current)
            </Typography>
            <TextInput
                source="password"
                label="New Password"
                type="password"
                validate={[minLength(8)]}
                fullWidth
            />
        </SimpleForm>
    </Edit>
);
