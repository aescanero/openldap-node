import React from 'react';
import {
    List,
    Datagrid,
    TextField,
    ArrayField,
    ChipField,
    SingleFieldList,
    EditButton,
    DeleteButton,
    SearchInput,
    CreateButton,
    TopToolbar,
    FilterButton,
    ExportButton,
} from 'react-admin';
import { Card, CardContent, Typography } from '@mui/material';

const GroupListActions = () => (
    <TopToolbar>
        <FilterButton />
        <CreateButton />
        <ExportButton />
    </TopToolbar>
);

const groupFilters = [
    <SearchInput key="q" source="q" alwaysOn placeholder="Search groups..." />,
];

const Empty = () => (
    <Card>
        <CardContent>
            <Typography variant="h6" paragraph>
                No groups yet
            </Typography>
            <Typography variant="body2">
                Create one by clicking the "Create" button
            </Typography>
        </CardContent>
    </Card>
);

export const GroupList = () => (
    <List
        filters={groupFilters}
        actions={<GroupListActions />}
        empty={<Empty />}
        perPage={20}
        sort={{ field: 'group_name', order: 'ASC' }}
    >
        <Datagrid rowClick="edit" bulkActionButtons={false}>
            <TextField source="group_name" label="Group Name" />
            <TextField source="description" label="Description" />
            <TextField source="gid_number" label="GID" />
            <ArrayField source="members" label="Members">
                <SingleFieldList linkType={false}>
                    <ChipField source="id" size="small" />
                </SingleFieldList>
            </ArrayField>
            <EditButton />
            <DeleteButton />
        </Datagrid>
    </List>
);
