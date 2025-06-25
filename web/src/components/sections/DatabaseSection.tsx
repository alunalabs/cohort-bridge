'use client';

import { useFormContext } from 'react-hook-form';
import { Database, Plus, Minus } from 'lucide-react';

interface DatabaseSectionProps {
    configType: string;
}

export default function DatabaseSection({ configType }: DatabaseSectionProps) {
    const { register, watch, setValue } = useFormContext();

    const databaseType = watch('database.type');
    const fields = watch('database.fields') || [];
    const isTokenized = watch('database.is_tokenized');

    const addField = () => {
        setValue('database.fields', [...fields, '']);
    };

    const removeField = (index: number) => {
        const newFields = fields.filter((_: any, i: number) => i !== index);
        setValue('database.fields', newFields);
    };

    const updateField = (index: number, value: string) => {
        const newFields = [...fields];
        newFields[index] = value;
        setValue('database.fields', newFields);
    };

    return (
        <div className="bg-white rounded-xl shadow-sm border border-slate-200 p-6">
            <div className="mb-6">
                <div className="flex items-center space-x-3 mb-2">
                    <div className="w-8 h-8 bg-blue-100 rounded-lg flex items-center justify-center">
                        <Database className="h-4 w-4 text-blue-600" />
                    </div>
                    <h3 className="text-lg font-semibold text-slate-900">Database Configuration</h3>
                </div>
                <p className="text-sm text-slate-600 ml-11">
                    Configure your data source and field mappings for privacy-preserving record linkage.
                </p>
            </div>

            <div className="space-y-6">
                {/* Database Type */}
                <div>
                    <label className="block text-sm font-medium text-slate-700 mb-3">
                        Database Type
                    </label>
                    <div className="grid grid-cols-2 gap-3">
                        <button
                            type="button"
                            onClick={() => setValue('database.type', 'csv')}
                            className={`p-4 rounded-lg border-2 transition-all duration-200 cursor-pointer ${databaseType === 'csv'
                                ? 'border-blue-500 bg-blue-50 text-blue-700'
                                : 'border-slate-300 hover:border-slate-400 text-slate-700'
                                }`}
                        >
                            <div className="text-center">
                                <div className="text-lg font-semibold">CSV File</div>
                                <div className="text-xs text-slate-500 mt-1">Comma-separated values</div>
                            </div>
                        </button>
                        <button
                            type="button"
                            onClick={() => setValue('database.type', 'postgres')}
                            className={`p-4 rounded-lg border-2 transition-all duration-200 cursor-pointer ${databaseType === 'postgres'
                                ? 'border-blue-500 bg-blue-50 text-blue-700'
                                : 'border-slate-300 hover:border-slate-400 text-slate-700'
                                }`}
                        >
                            <div className="text-center">
                                <div className="text-lg font-semibold">PostgreSQL</div>
                                <div className="text-xs text-slate-500 mt-1">Database server</div>
                            </div>
                        </button>
                    </div>
                </div>

                {/* Tokenized Toggle for tokenized config type */}
                {configType === 'tokenized' && (
                    <div>
                        <label className="flex items-center space-x-3">
                            <input
                                type="checkbox"
                                {...register('database.is_tokenized')}
                                className="rounded border-slate-300 text-blue-600 focus:ring-blue-500"
                            />
                            <span className="text-sm font-medium text-slate-700">Use tokenized data</span>
                        </label>
                    </div>
                )}

                {/* CSV File Configuration */}
                {databaseType === 'csv' && !isTokenized && (
                    <div>
                        <label className="block text-sm font-medium text-slate-700 mb-2">
                            CSV Filename <span className="text-red-500">*</span>
                        </label>
                        <input
                            type="text"
                            {...register('database.filename')}
                            placeholder="data/patients.csv"
                            className="w-full px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 text-slate-900 placeholder-slate-400 bg-white"
                        />
                    </div>
                )}

                {/* Tokenized File Configuration */}
                {isTokenized && (
                    <div>
                        <label className="block text-sm font-medium text-slate-700 mb-2">
                            Tokenized File Path
                        </label>
                        <input
                            type="text"
                            {...register('database.tokenized_file')}
                            placeholder="out/tokens_party_a.json"
                            className="w-full px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 text-slate-900 placeholder-slate-400 bg-white"
                        />
                    </div>
                )}

                {/* PostgreSQL Configuration */}
                {databaseType === 'postgres' && (
                    <div className="space-y-4">
                        <div className="grid grid-cols-2 gap-4">
                            <div>
                                <label className="block text-sm font-medium text-slate-700 mb-2">
                                    Host <span className="text-red-500">*</span>
                                </label>
                                <input
                                    type="text"
                                    {...register('database.host')}
                                    placeholder="localhost"
                                    className="w-full px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 text-slate-900 placeholder-slate-400 bg-white"
                                />
                            </div>
                            <div>
                                <label className="block text-sm font-medium text-slate-700 mb-2">
                                    Port <span className="text-red-500">*</span>
                                </label>
                                <input
                                    type="number"
                                    {...register('database.port', { valueAsNumber: true })}
                                    placeholder="5432"
                                    className="w-full px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 text-slate-900 placeholder-slate-400 bg-white"
                                />
                            </div>
                        </div>

                        <div className="grid grid-cols-2 gap-4">
                            <div>
                                <label className="block text-sm font-medium text-slate-700 mb-2">
                                    Username <span className="text-red-500">*</span>
                                </label>
                                <input
                                    type="text"
                                    {...register('database.user')}
                                    placeholder="cohort_user"
                                    className="w-full px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 text-slate-900 placeholder-slate-400 bg-white"
                                />
                            </div>
                            <div>
                                <label className="block text-sm font-medium text-slate-700 mb-2">
                                    Password
                                </label>
                                <input
                                    type="password"
                                    {...register('database.password')}
                                    placeholder="your_password_here"
                                    className="w-full px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 text-slate-900 placeholder-slate-400 bg-white"
                                />
                            </div>
                        </div>

                        <div className="grid grid-cols-2 gap-4">
                            <div>
                                <label className="block text-sm font-medium text-slate-700 mb-2">
                                    Database Name <span className="text-red-500">*</span>
                                </label>
                                <input
                                    type="text"
                                    {...register('database.dbname')}
                                    placeholder="cohort_database"
                                    className="w-full px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 text-slate-900 placeholder-slate-400 bg-white"
                                />
                            </div>
                            <div>
                                <label className="block text-sm font-medium text-slate-700 mb-2">
                                    Table Name <span className="text-red-500">*</span>
                                </label>
                                <input
                                    type="text"
                                    {...register('database.table')}
                                    placeholder="users"
                                    className="w-full px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 text-slate-900 placeholder-slate-400 bg-white"
                                />
                            </div>
                        </div>
                    </div>
                )}

                {/* Fields Configuration */}
                {!isTokenized && (
                    <div>
                        <div className="flex items-center justify-between mb-3">
                            <div>
                                <label className="block text-sm font-medium text-slate-700">
                                    Database Field Mapping
                                </label>
                                <p className="text-sm text-slate-600 mt-1">
                                    Map your database column names to data types for record linkage
                                </p>
                            </div>
                            <button
                                type="button"
                                onClick={addField}
                                className="flex items-center space-x-1 text-blue-600 hover:text-blue-700 text-sm font-medium"
                            >
                                <Plus className="h-4 w-4" />
                                <span>Add Field</span>
                            </button>
                        </div>

                        <div className="space-y-3">
                            {fields.map((field: string, index: number) => {
                                const fieldLabels = [
                                    'First Name Column',
                                    'Last Name Column',
                                    'Date of Birth Column',
                                    'ZIP Code Column',
                                    'Email Column',
                                    'ID Column',
                                    'Phone Column',
                                    'Address Column'
                                ];

                                return (
                                    <div key={index} className="flex items-center space-x-3">
                                        <div className="flex-1">
                                            <label className="block text-xs font-medium text-slate-600 mb-1">
                                                {fieldLabels[index] || `Field ${index + 1} Column`}
                                            </label>
                                            <input
                                                type="text"
                                                value={field}
                                                onChange={(e) => updateField(index, e.target.value)}
                                                placeholder={`${fieldLabels[index]?.split(' ')[0].toLowerCase() || 'field'}_name`}
                                                className="w-full px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                                            />
                                        </div>
                                        {fields.length > 1 && (
                                            <button
                                                type="button"
                                                onClick={() => removeField(index)}
                                                className="mt-6 p-2 text-red-600 hover:text-red-700 hover:bg-red-50 rounded-lg transition-colors"
                                            >
                                                <Minus className="h-4 w-4" />
                                            </button>
                                        )}
                                    </div>
                                );
                            })}
                        </div>
                    </div>
                )}

                {/* Random Bits Percent */}
                <div>
                    <label className="block text-sm font-medium text-slate-700 mb-2">
                        Random Bits Percentage (0.0 - 1.0)
                    </label>
                    <input
                        type="number"
                        step="0.01"
                        min="0"
                        max="1"
                        {...register('database.random_bits_percent', { valueAsNumber: true })}
                        placeholder="0.0"
                        className="w-full px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                    />
                    <p className="mt-1 text-xs text-slate-600">
                        Higher values (0.05-0.1) increase privacy but reduce matching accuracy. Use 0.0 for maximum accuracy.
                    </p>
                </div>
            </div>
        </div>
    );
} 