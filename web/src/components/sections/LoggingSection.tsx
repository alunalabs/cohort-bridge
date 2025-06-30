'use client';

import { useFormContext } from 'react-hook-form';
import { FileText } from 'lucide-react';

interface LoggingSectionProps {
    missingFields?: string[];
}

export default function LoggingSection({ missingFields = [] }: LoggingSectionProps) {
    const { register, setValue, watch } = useFormContext();

    const logLevel = watch('logging.level');

    const getInputClass = (fieldName: string) => {
        const baseClass = "w-full px-3 py-2 border rounded-lg focus:ring-2 text-slate-900 placeholder-slate-400 bg-white transition-colors";
        const isMissing = missingFields.includes(fieldName);

        if (isMissing) {
            return `${baseClass} border-red-300 bg-red-50 focus:ring-red-500 focus:border-red-500`;
        }

        return `${baseClass} border-slate-300 focus:ring-blue-500 focus:border-blue-500`;
    };

    return (
        <div className="bg-white rounded-xl shadow-sm border border-slate-200 p-6">
            <div className="mb-6">
                <div className="flex items-center space-x-3 mb-2">
                    <div className="w-8 h-8 bg-indigo-100 rounded-lg flex items-center justify-center">
                        <FileText className="h-4 w-4 text-indigo-600" />
                    </div>
                    <h3 className="text-lg font-semibold text-slate-900">Logging Configuration</h3>
                </div>
                <p className="text-sm text-slate-600 ml-11">
                    Configure logging levels and output destinations for monitoring and debugging record linkage operations.
                </p>
            </div>

            <div className="space-y-6">
                {/* Log Level */}
                <div>
                    <label className="block text-sm font-medium text-slate-700 mb-3">
                        Log Level
                    </label>
                    <div className="grid grid-cols-2 gap-3">
                        {['info', 'debug', 'warn', 'error'].map((level) => (
                            <button
                                key={level}
                                type="button"
                                onClick={() => setValue('logging.level', level)}
                                className={`p-3 rounded-lg border-2 transition-all duration-200 cursor-pointer ${logLevel === level
                                    ? 'border-blue-500 bg-blue-50 text-blue-700'
                                    : 'border-slate-300 hover:border-slate-400 text-slate-700'
                                    }`}
                            >
                                <div className="text-center">
                                    <div className="text-sm font-semibold capitalize">{level}</div>
                                    <div className="text-xs text-slate-500 mt-1">
                                        {level === 'info' && 'General information'}
                                        {level === 'debug' && 'Detailed debugging'}
                                        {level === 'warn' && 'Warning messages'}
                                        {level === 'error' && 'Error messages only'}
                                    </div>
                                </div>
                            </button>
                        ))}
                    </div>
                    <p className="mt-2 text-xs text-slate-600">
                        <strong>Info:</strong> Standard operations, <strong>Debug:</strong> Detailed tracing,
                        <strong>Warn:</strong> Potential issues, <strong>Error:</strong> Critical problems only
                    </p>
                </div>

                {/* Log File Path */}
                <div>
                    <label className="block text-sm font-medium text-slate-700 mb-2">
                        Log File Path
                    </label>
                    <input
                        type="text"
                        {...register('logging.file_path')}
                        placeholder="logs/cohort-bridge.log"
                        className={getInputClass('logging.file_path')}
                    />
                    <p className="mt-1 text-xs text-slate-600">
                        Path where log files will be written. Directory will be created if it doesn't exist.
                    </p>
                </div>

                {/* Console Logging */}
                <div>
                    <label className="flex items-center space-x-3">
                        <input
                            type="checkbox"
                            {...register('logging.console_enabled')}
                            className="rounded border-slate-300 text-blue-600 focus:ring-blue-500"
                        />
                        <span className="text-sm font-medium text-slate-700">Enable Console Logging</span>
                    </label>
                    <p className="mt-1 text-xs text-slate-600 ml-6">
                        Output logs to console/terminal in addition to file. Useful for development and debugging.
                    </p>
                </div>

                {/* Log Rotation */}
                <div>
                    <label className="flex items-center space-x-3">
                        <input
                            type="checkbox"
                            {...register('logging.rotation_enabled')}
                            className="rounded border-slate-300 text-blue-600 focus:ring-blue-500"
                        />
                        <span className="text-sm font-medium text-slate-700">Enable Log Rotation</span>
                    </label>
                    <p className="mt-1 text-xs text-slate-600 ml-6">
                        Automatically rotate log files when they reach a certain size to prevent disk space issues.
                    </p>
                </div>

                {/* Max File Size */}
                <div>
                    <label className="block text-sm font-medium text-slate-700 mb-2">
                        Max File Size (MB)
                    </label>
                    <input
                        type="number"
                        {...register('logging.max_file_size_mb', { valueAsNumber: true })}
                        placeholder="100"
                        min="1"
                        max="1000"
                        className={getInputClass('logging.max_file_size_mb')}
                    />
                    <p className="mt-1 text-xs text-slate-600">
                        Maximum size of each log file before rotation. 50-200MB recommended.
                    </p>
                </div>

                {/* Max Files */}
                <div>
                    <label className="block text-sm font-medium text-slate-700 mb-2">
                        Max Backup Files
                    </label>
                    <input
                        type="number"
                        {...register('logging.max_backup_files', { valueAsNumber: true })}
                        placeholder="5"
                        min="1"
                        max="50"
                        className={getInputClass('logging.max_backup_files')}
                    />
                    <p className="mt-1 text-xs text-slate-600">
                        Number of rotated log files to keep. Older files are automatically deleted.
                    </p>
                </div>
            </div>
        </div>
    );
} 