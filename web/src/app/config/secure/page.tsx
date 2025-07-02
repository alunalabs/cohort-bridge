'use client';

import { Shield } from 'lucide-react';
import FlexibleConfigBuilder from '@/components/FlexibleConfigBuilder';

export default function SecureConfigPage() {
    const defaultValues = {
        database: {
            type: 'csv',
            filename: 'data/patients_secure.csv',
            fields: ['name:FIRST', 'name:LAST', 'date:BIRTHDATE', 'gender:GENDER', 'zip:ZIP'],
            random_bits_percent: 0.05,
        },
        peer: {
            host: '192.168.1.100',
            port: 8081,
        },
        listen_port: 8082,

        security: {
            rate_limit_per_min: 3,
        },
        timeouts: {
            connection_timeout: '45s',
            read_timeout: '120s',
            write_timeout: '120s',
            idle_timeout: '600s',
            handshake_timeout: '60s',
        },
        logging: {
            level: 'info',
            file: 'logs/cohort_secure.log',
            max_size: 50,
            max_backups: 5,
            max_age: 7,
            enable_syslog: false,
            enable_audit: true,
            audit_file: 'logs/security_audit.log',
        },
    };

    return (
        <FlexibleConfigBuilder
            title="Secure Configuration"
            description="Maximum security with comprehensive logging, audit trails, and data normalization"
            defaultSections={['database', 'peer', 'security', 'timeouts', 'logging']}
            defaultValues={defaultValues}
            icon={Shield}
        />
    );
} 