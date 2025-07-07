'use client';

import { Database } from 'lucide-react';
import FlexibleConfigBuilder from '@/components/FlexibleConfigBuilder';

export default function PostgresConfigPage() {
    const defaultValues = {
        database: {
            type: 'postgres',
            host: 'localhost',
            port: 5432,
            user: 'cohort_user',
            password: '',
            dbname: 'cohort_database',
            table: 'users',
            fields: ['id', 'name:first_name', 'name:last_name', 'date:date_of_birth', 'gender:gender', 'zip:zip_code', 'email'],
            random_bits_percent: 0.0,
        },
        peer: {
            host: 'localhost',
            port: 8080,
        },
        listen_port: 8080,
    };

    return (
        <FlexibleConfigBuilder
            title="PostgreSQL Configuration"
            description="Connect to PostgreSQL databases with enhanced security and data normalization"
            defaultSections={['database', 'peer']}
            defaultValues={defaultValues}
            icon={Database}
        />
    );
} 