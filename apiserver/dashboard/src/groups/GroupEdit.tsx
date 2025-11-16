import React from 'react';
import {
    Edit,
    SimpleForm,
    TextInput,
    ArrayInput,
    SimpleFormIterator,
    required,
    DeleteButton,
    TopToolbar,
    ListButton,
} from 'react-admin';
import { Typography } from '@mui/material';

const GroupEditActions = () => (
    <TopToolbar>
        <ListButton />
        <DeleteButton />
    </TopToolbar>
);

export const GroupEdit = () => (
    <Edit actions={<GroupEditActions />}>
        <SimpleForm>
            <Typography variant="h6" gutterBottom>
                Edit Group
            </Typography>
            <TextInput
                source="group_name"
                label="Group Name"
                disabled
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
    </Edit>
);
