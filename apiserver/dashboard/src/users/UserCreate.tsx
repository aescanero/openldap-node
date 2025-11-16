import React from 'react';
import {
    Create,
    SimpleForm,
    TextInput,
    email,
    required,
    minLength,
} from 'react-admin';
import { Box, Typography } from '@mui/material';

export const UserCreate = () => (
    <Create redirect="list">
        <SimpleForm>
            <Typography variant="h6" gutterBottom>
                Create New User
            </Typography>
            <Box display="flex" gap={2} width="100%">
                <TextInput
                    source="username"
                    label="Username"
                    validate={[required(), minLength(3)]}
                    fullWidth
                />
                <TextInput
                    source="password"
                    label="Password"
                    type="password"
                    validate={[required(), minLength(8)]}
                    fullWidth
                />
            </Box>
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
        </SimpleForm>
    </Create>
);
