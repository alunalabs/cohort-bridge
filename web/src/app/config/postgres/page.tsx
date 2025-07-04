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

        security: {
            rate_limit_per_min: 5,
        },
        timeouts: {
            connection_timeout: '30s',
            read_timeout: '60s',
            write_timeout: '60s',
            idle_timeout: '300s',
            handshake_timeout: '30s',
        },
    };

    return (
        <FlexibleConfigBuilder
            title="PostgreSQL Configuration"
            description="Connect to PostgreSQL databases with enhanced security and data normalization"
            defaultSections={['database', 'peer', 'security', 'timeouts']}
            defaultValues={defaultValues}
            icon={Database}
        />
    );
} 