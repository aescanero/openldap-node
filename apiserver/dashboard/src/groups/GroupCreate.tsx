import React from 'react';
import {
    Create,
    SimpleForm,
    TextInput,
    ArrayInput,
    SimpleFormIterator,
    required,
    minLength,
} from 'react-admin';
import { Box, Typography } from '@mui/material';

export const GroupCreate = () => (
    <Create redirect="list">
        <SimpleForm>
            <Typography variant="h6" gutterBottom>
                Create New Group
            </Typography>
            <TextInput
                source="group_name"
                label="Group Name"
                validate={[required(), minLength(2)]}
                fullWidth
            />
            <TextInput
                source="description"
                label="Description"
                multiline
                rows={3}
                fullWidth
            />
            <Typography variant="subtitle1" gutterBottom sx={{ mt: 2 }}>
                Members (usernames)
            </Typography>
            <ArrayInput source="members" label="">
                <SimpleFormIterator inline>
                    <TextInput source="" label="Username" helperText={false} />
                </SimpleFormIterator>
            </ArrayInput>
        </SimpleForm>
    </Create>
);
