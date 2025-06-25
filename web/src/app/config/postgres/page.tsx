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
            fields: ['id', 'first_name', 'last_name', 'email', 'date_of_birth'],
            random_bits_percent: 0.0,
        },
        peer: {
            host: 'localhost',
            port: 8081,
        },
        listen_port: 8080,
        private_key: '',
        security: {
            allowed_ips: ['127.0.0.1', '::1', '192.168.1.0/24'],
            require_ip_check: true,
            max_connections: 10,
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
            description="Connect to PostgreSQL databases with enhanced security"
            defaultSections={['database', 'peer', 'security', 'timeouts']}
            defaultValues={defaultValues}
            icon={Database}
        />
    );
} 