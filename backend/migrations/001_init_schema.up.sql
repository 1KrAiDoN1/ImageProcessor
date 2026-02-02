-- Create images table
CREATE TABLE IF NOT EXISTS images (
    id VARCHAR(36) PRIMARY KEY,
    original_filename VARCHAR(255) NOT NULL,
    original_size BIGINT NOT NULL,
    mime_type VARCHAR(100) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'uploaded',
    original_path VARCHAR(500) NOT NULL,
    bucket VARCHAR(100) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create index for status queries
CREATE INDEX idx_images_status ON images(status);
CREATE INDEX idx_images_created_at ON images(created_at DESC);

-- Create processed_images table
CREATE TABLE IF NOT EXISTS processed_images (
    id VARCHAR(36) PRIMARY KEY,
    image_id VARCHAR(36) NOT NULL,
    operation VARCHAR(50) NOT NULL,
    parameters JSONB,
    path VARCHAR(500) NOT NULL,
    size BIGINT NOT NULL,
    mime_type VARCHAR(100) NOT NULL,
    format VARCHAR(20) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'processing',
    error_message TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (image_id) REFERENCES images(id) ON DELETE CASCADE
);

-- Create indexes for processed_images
CREATE INDEX idx_processed_images_image_id ON processed_images(image_id);
CREATE INDEX idx_processed_images_operation ON processed_images(operation);
CREATE INDEX idx_processed_images_status ON processed_images(status);

-- Create processing_jobs table (для отслеживания задач в Kafka)
CREATE TABLE IF NOT EXISTS processing_jobs (
    id VARCHAR(36) PRIMARY KEY,
    image_id VARCHAR(36) NOT NULL,
    operations JSONB NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    attempts INT NOT NULL DEFAULT 0,
    max_attempts INT NOT NULL DEFAULT 3,
    error_message TEXT,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (image_id) REFERENCES images(id) ON DELETE CASCADE
);

-- Create indexes for processing_jobs
CREATE INDEX idx_processing_jobs_image_id ON processing_jobs(image_id);
CREATE INDEX idx_processing_jobs_status ON processing_jobs(status);
CREATE INDEX idx_processing_jobs_created_at ON processing_jobs(created_at DESC);

-- Create statistics table
CREATE TABLE IF NOT EXISTS statistics (
    id VARCHAR(36) PRIMARY KEY,
    total_images_uploaded BIGINT NOT NULL DEFAULT 0,
    total_images_processed BIGINT NOT NULL DEFAULT 0,
    total_images_failed BIGINT NOT NULL DEFAULT 0,
    total_data_processed_bytes BIGINT NOT NULL DEFAULT 0,
    average_processing_time_ms NUMERIC(10, 2) NOT NULL DEFAULT 0,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create operation_statistics table (для детальной статистики по операциям)
CREATE TABLE IF NOT EXISTS operation_statistics (
    id SERIAL PRIMARY KEY,
    operation_type VARCHAR(50) NOT NULL,
    total_count BIGINT NOT NULL DEFAULT 0,
    success_count BIGINT NOT NULL DEFAULT 0,
    failure_count BIGINT NOT NULL DEFAULT 0,
    average_processing_time_ms NUMERIC(10, 2) NOT NULL DEFAULT 0,
    total_processing_time_ms BIGINT NOT NULL DEFAULT 0,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(operation_type)
);

-- Insert default statistics row
INSERT INTO statistics (id, total_images_uploaded, total_images_processed, total_images_failed, total_data_processed_bytes, average_processing_time_ms)
VALUES ('default', 0, 0, 0, 0, 0)
ON CONFLICT (id) DO NOTHING;

-- Create function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers for updated_at
CREATE TRIGGER update_images_updated_at BEFORE UPDATE ON images
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_processing_jobs_updated_at BEFORE UPDATE ON processing_jobs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_statistics_updated_at BEFORE UPDATE ON statistics
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_operation_statistics_updated_at BEFORE UPDATE ON operation_statistics
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

